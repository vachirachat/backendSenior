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
func (handler *RoomRouteHandler) Mount(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/:id/member", handler.getRoomMember)
	routerGroup.POST("/:id/member", handler.addMemberToRoom)
	routerGroup.DELETE("/:id/member", handler.deleteMemberFromRoom)
	routerGroup.GET("/:id/proxy", handler.getRoomProxies)
	routerGroup.POST("/:id/proxy", handler.addProxiesToRoom)
	routerGroup.DELETE("/:id/proxy", handler.removeProxiesFromRoom)

	routerGroup.POST("/", handler.authMw.AuthRequired(), handler.addRoomHandler)
	routerGroup.POST("/:id/name" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.editRoomNameHandler)
	routerGroup.DELETE("/:id" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.deleteRoomByIDHandler)
	// routerGroup.POST("/addmembertoroom" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.addMemberToRoom)
	// routerGroup.POST("/deletemembertoroom" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.deleteMemberFromRoom)
	routerGroup.GET("/:id" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.getRoomByIDHandler)
	routerGroup.GET("/", handler.authMw.AuthRequired(), handler.roomListHandler)
}

func (handler *RoomRouteHandler) roomListHandler(context *gin.Context) {
	var roomsInfo model.RoomInfo

	isMe := context.Query("me") != ""

	var rooms []model.Room
	var err error

	if isMe {
		myID := context.GetString(auth.UserIdField)
		rooms, err = handler.roomService.GetUserRooms(myID)
	} else {
		rooms, err = handler.roomService.GetAllRooms()
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
func (handler *RoomRouteHandler) getRoomByIDHandler(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}

	room, err := handler.roomService.GetRoomByID(roomID)
	if err != nil {
		log.Println("error GetRoomByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, room)
}

// create an empty room, then the creator of the room is automatically invited to the room
func (handler *RoomRouteHandler) addRoomHandler(context *gin.Context) {
	var room model.Room
	var roomID string
	err := context.BindJSON(&room)
	isOK := false

	defer func() {
		if !isOK && roomID != "" {
			handler.roomService.DeleteRoomByID(roomID)
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

	// check org existence
	_, err = handler.orgService.GetOrganizeById(room.OrgID.Hex())
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error, room might not exist"})
		return
	}

	roomID, err = handler.roomService.AddRoom(room)
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	// Invite self to rooms
	userID := context.GetString(auth.UserIdField)
	err = handler.roomService.AddMembersToRoom(roomID, []string{userID})
	if err != nil {
		log.Println("error AddRoomHandeler; invite self to room", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	// Add room to org
	err = handler.orgService.AddRoomsToOrg(room.OrgID.Hex(), []string{roomID})
	if err != nil {
		log.Println("error AddRoomHandeler; invite room to org", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	isOK = true
	context.JSON(http.StatusCreated, gin.H{"status": "success", "roomId": roomID})
}

func (handler *RoomRouteHandler) editRoomNameHandler(context *gin.Context) {
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

	err = handler.roomService.EditRoomName(roomID, room)

	if err != nil {
		log.Println("error EditRoomNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *RoomRouteHandler) deleteRoomByIDHandler(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}

	err := handler.roomService.DeleteRoomByID(roomID)
	if err != nil {
		log.Println("error DeleteRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Match with Socket-structure

//// -- JoinAPI -> getSession(Topic+#ID) -> giveUserSession
func (handler *RoomRouteHandler) addMemberToRoom(context *gin.Context) {
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

	err = handler.roomService.AddMembersToRoom(roomID, userIDs)
	if err != nil {
		log.Println("error AddMemberToRoom", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	// send event to proxy
	err = handler.chatService.BroadcastMessageToRoom(roomID, chatsocket.Message{
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

func (handler *RoomRouteHandler) deleteMemberFromRoom(context *gin.Context) {
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

	err = handler.roomService.DeleteMemberFromRoom(roomID, userIDs)

	if err != nil {
		log.Println("error DeleteRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	// send event to proxy
	err = handler.chatService.BroadcastMessageToRoom(roomID, chatsocket.Message{
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
func (handler *RoomRouteHandler) getRoomMember(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}
	isFull := context.Query("full") != ""

	userIDs, err := handler.roomService.GetRoomMemberIDs(roomID)
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
	users, err := handler.userService.GetUsersByIDs(userIDs)
	if err != nil {
		fmt.Println("[getRoomMember, full] error", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"members": users,
	})
}

func (handler *RoomRouteHandler) getRoomProxies(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}

	proxyIDs, err := handler.roomService.GetRoomProxyIDs(roomID)
	if err != nil {
		fmt.Println("[getRoomProxies]", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	proxies, err := handler.proxyService.GetProxiesByIDs(proxyIDs)
	if err != nil {
		fmt.Println("[getRoomProxies]", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"proxies": proxies,
	})

}

func (handler *RoomRouteHandler) addProxiesToRoom(context *gin.Context) {
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

	err := handler.roomService.AddProxiesToRoom(roomID, utills.ToStringArr(body.ProxyIDs))
	if err != nil {
		fmt.Println("[Room handler] addProxiesToRoom: ", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})

}

func (handler *RoomRouteHandler) removeProxiesFromRoom(context *gin.Context) {
	roomID := context.Param("id")

	var body struct {
		ProxyIDs []bson.ObjectId `json:"proxyIds"`
	}

	if err := context.MustBindWith(&body, binding.JSON); err != nil {
		return
	}

	err := handler.roomService.DeleteProxiesFromRoom(roomID, utills.ToStringArr(body.ProxyIDs))
	if err != nil {
		fmt.Println("[Room handler] addProxiesToRoom: ", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})

}
