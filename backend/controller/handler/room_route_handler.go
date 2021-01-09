package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/message_types"
	"backendSenior/domain/model/chatsocket/room"
	"backendSenior/domain/service"
	"backendSenior/utills"
	"errors"
	"fmt"

	"backendSenior/domain/model"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/globalsign/mgo/bson"
)

type RoomRouteHandler struct {
	roomService  *service.RoomService
	userService  *service.UserService  // for get room members route
	proxyService *service.ProxyService // for get room proxies route
	authMw       *auth.JWTMiddleware
	chatService  *service.ChatService     // for broadcast event when join/leave room
	orgService   *service.OrganizeService // for add room to org
}

// NewRoomHandler create new handler for room
func NewRoomRouteHandler(roomService *service.RoomService, authMw *auth.JWTMiddleware, userService *service.UserService, proxyService *service.ProxyService, chatService *service.ChatService, orgService *service.OrganizeService) *RoomRouteHandler {
	return &RoomRouteHandler{
		roomService:  roomService,
		userService:  userService,
		authMw:       authMw,
		proxyService: proxyService,
		chatService:  chatService,
		orgService:   orgService,
	}
}

//Mount make RoomRouteHandler handler request from specific `RouterGroup`
func (h *RoomRouteHandler) Mount(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/:id/member", h.getRoomMember)
	routerGroup.POST("/:id/member", h.addMemberToRoom)
	routerGroup.DELETE("/:id/member", h.deleteMemberFromRoom)
	routerGroup.GET("/:id/proxy", h.getRoomProxies)
	routerGroup.POST("/:id/proxy", h.addProxiesToRoom)
	routerGroup.DELETE("/:id/proxy", h.removeProxiesFromRoom)

	routerGroup.POST("/:id/master-proxy", h.setMasterProxy)
	routerGroup.GET("/:id/master-proxy", h.getMasterProxy)

	routerGroup.POST("/", h.authMw.AuthRequired(), h.addRoomHandler)
	routerGroup.POST("/:id/name" /*h.authService.AuthMiddleware("object", "view"),*/, h.editRoomNameHandler)
	routerGroup.DELETE("/:id" /*h.authService.AuthMiddleware("object", "view"),*/, h.deleteRoomByIDHandler)
	// routerGroup.POST("/addmembertoroom" /*h.authService.AuthMiddleware("object", "view"),*/, h.addMemberToRoom)
	// routerGroup.POST("/deletemembertoroom" /*h.authService.AuthMiddleware("object", "view"),*/, h.deleteMemberFromRoom)
	routerGroup.GET("/:id" /*h.authService.AuthMiddleware("object", "view"),*/, h.getRoomByIDHandler)
	routerGroup.GET("/", h.authMw.AuthRequired(), h.roomListHandler)
}

func (h *RoomRouteHandler) roomListHandler(context *gin.Context) {
	var roomsInfo model.RoomInfo

	isMe := context.Query("me") != ""

	var rooms []model.Room
	var err error

	if isMe {
		myID := context.GetString(auth.UserIdField)
		rooms, err = h.roomService.GetUserRooms(myID)
	} else {
		rooms, err = h.roomService.GetAllRooms()
	}
	if err != nil {
		log.Println("error roomListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	roomsInfo.Room = rooms
	context.JSON(http.StatusOK, roomsInfo)
}

// for get room by id
func (h *RoomRouteHandler) getRoomByIDHandler(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}

	room, err := h.roomService.GetRoomByID(roomID)
	if err != nil {
		log.Println("error GetRoomByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, room)
}

// create an empty room, then the creator of the room is automatically invited to the room
func (h *RoomRouteHandler) addRoomHandler(context *gin.Context) {
	var room model.Room
	var roomID string
	err := context.BindJSON(&room)
	isOK := false

	defer func() {
		if !isOK && roomID != "" {
			h.roomService.DeleteRoomByID(roomID)
		}
	}()

	if err != nil || room.OrgID.Hex() == "" {
		if err == nil {
			err = errors.New("org ID is required")
		}
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	orgID := room.OrgID.Hex()
	room.OrgID = "" // when invite room to org requires that room has no org!
	// check org existence
	_, err = h.orgService.GetOrganizeById(orgID)
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error, room might not exist"})
		return
	}

	roomID, err = h.roomService.AddRoom(room)
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	// Invite self to rooms
	userID := context.GetString(auth.UserIdField)
	err = h.roomService.AddMembersToRoom(roomID, []string{userID})
	if err != nil {
		log.Println("error AddRoomHandeler; invite self to room", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	// Add room to org
	err = h.orgService.AddRoomsToOrg(orgID, []string{roomID})
	if err != nil {
		log.Println("error AddRoomHandeler; invite room to org", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	isOK = true
	context.JSON(http.StatusCreated, gin.H{"status": "success", "roomId": roomID})
}

func (h *RoomRouteHandler) editRoomNameHandler(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}

	var room model.Room
	err := context.ShouldBindJSON(&room)
	room.RoomID = ""

	if err != nil {
		log.Println("error EditRoomNametHandler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	err = h.roomService.EditRoomName(roomID, room)

	if err != nil {
		log.Println("error EditRoomNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *RoomRouteHandler) deleteRoomByIDHandler(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}

	room, err := h.roomService.GetRoomByID(roomID)
	if err != nil {
		log.Println("error DeleteRoomHandler before deleting room", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	userIDs := utills.ToStringArr(room.ListUser)
	proxyIDs := utills.ToStringArr(room.ListProxy)
	orgID := room.OrgID.Hex()

	// TODO: make it transaction
	// delete room-user relation
	err = h.roomService.DeleteMemberFromRoom(roomID, userIDs)
	if err != nil {
		log.Println("error DeleteRoomHandler removing members from room", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}
	err = h.roomService.DeleteProxiesFromRoom(roomID, proxyIDs)
	if err != nil {
		log.Println("error DeleteRoomHandler removing proxies from room", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}
	err = h.orgService.DeleteRoomsFromOrg(orgID, []string{roomID})
	if err != nil {
		log.Println("error DeleteRoomHandler removing room from org", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	err = h.roomService.DeleteRoomByID(roomID)
	if err != nil {
		log.Println("error DeleteRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Match with Socket-structure

//// -- JoinAPI -> getSession(Topic+#ID) -> giveUserSession
func (h *RoomRouteHandler) addMemberToRoom(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}

	// use bson.ObjectID to validate when bind
	var body struct {
		UserIDs []bson.ObjectId `json:"userIDs"`
	}

	err := context.ShouldBindJSON(&body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	userIDs := utills.ToStringArr(body.UserIDs)

	err = h.roomService.AddMembersToRoom(roomID, userIDs)
	if err != nil {
		log.Println("error AddMemberToRoom", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	// send event to proxy
	err = h.chatService.BroadcastMessageToRoom(roomID, chatsocket.Message{
		Type: message_types.Room,
		Payload: room.MemberEvent{
			Type:    room.Join,
			RoomID:  roomID,
			Members: userIDs,
		},
	})
	// this just print warning, the request is successful anyway
	if err != nil {
		fmt.Println("error bcast event to proxy", err)
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *RoomRouteHandler) deleteMemberFromRoom(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}

	// use bson.ObjectID to validate when bind
	var body struct {
		UserIDs []bson.ObjectId `json:"userIDs"`
	}

	err := context.ShouldBindJSON(&body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	userIDs := utills.ToStringArr(body.UserIDs)

	err = h.roomService.DeleteMemberFromRoom(roomID, userIDs)

	if err != nil {
		log.Println("error DeleteRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	// send event to proxy
	err = h.chatService.BroadcastMessageToRoom(roomID, chatsocket.Message{
		Type: message_types.Room,
		Payload: room.MemberEvent{
			Type:    room.Leave,
			RoomID:  roomID,
			Members: userIDs,
		},
	})
	// this just print warning, the request is successful anyway
	if err != nil {
		fmt.Println("error bcast event to proxy", err)
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

// TODO change this to return full object only, currently keep for compatibility of proxy
func (h *RoomRouteHandler) getRoomMember(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}
	isFull := context.Query("full") != ""

	userIDs, err := h.roomService.GetRoomMemberIDs(roomID)
	if err != nil {
		fmt.Println("[getRoomMember] error", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	if !isFull {
		context.JSON(http.StatusOK, gin.H{
			"members": userIDs,
		})
		return
	}
	users, err := h.userService.GetUsersByIDs(userIDs)
	if err != nil {
		fmt.Println("[getRoomMember, full] error", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"members": users,
	})
}

func (h *RoomRouteHandler) getRoomProxies(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}

	proxyIDs, err := h.roomService.GetRoomProxyIDs(roomID)
	if err != nil {
		fmt.Println("[getRoomProxies] get room proxy", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	proxies, err := h.proxyService.GetProxiesByIDs(proxyIDs)
	if err != nil {
		fmt.Println("[getRoomProxies] find proxy", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"proxies": proxies,
	})

}

func (h *RoomRouteHandler) addProxiesToRoom(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}

	var body struct {
		ProxyIDs []bson.ObjectId `json:"proxyIds"`
	}

	if err := context.MustBindWith(&body, binding.JSON); err != nil {
		return
	}

	err := h.roomService.AddProxiesToRoom(roomID, utills.ToStringArr(body.ProxyIDs))
	if err != nil {
		fmt.Println("[Room handler] addProxiesToRoom: ", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})

}

func (h *RoomRouteHandler) removeProxiesFromRoom(context *gin.Context) {
	roomID := context.Param("id")

	var body struct {
		ProxyIDs []bson.ObjectId `json:"proxyIds"`
	}

	if err := context.MustBindWith(&body, binding.JSON); err != nil {
		return
	}

	err := h.roomService.DeleteProxiesFromRoom(roomID, utills.ToStringArr(body.ProxyIDs))
	if err != nil {
		fmt.Println("[Room handler] addProxiesToRoom: ", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})

}

func (h *RoomRouteHandler) setMasterProxy(c *gin.Context) {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "Bad room ID"})
		return
	}

	var b struct {
		ProxyID string
	}
	err := c.ShouldBindJSON(&b)
	if b.ProxyID == "" || !bson.IsObjectIdHex(b.ProxyID) {
		c.JSON(400, gin.H{"status": "Bad or empty proxyID"})
		return
	}

	err = h.roomService.SetRoomMasterProxy(roomID, b.ProxyID)
	if err != nil {
		fmt.Println("[room/setMasterProxy] error:", err)
		c.JSON(500, gin.H{"status": "error"})
	}
	c.JSON(200, gin.H{"status": "OK"})
}

// return object: master proxy of the room
func (h *RoomRouteHandler) getMasterProxy(c *gin.Context) {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "Bad room ID"})
		return
	}

	mainProxy, err := h.roomService.GetRoomMasterProxy(roomID)
	if err != nil {
		fmt.Println("room/getMasterProxy", err)
		c.JSON(500, gin.H{"status": "error"})
		return
	}
	c.JSON(200, mainProxy)
}
