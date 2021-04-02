package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/message_types"
	"backendSenior/domain/model/chatsocket/room"
	"backendSenior/domain/service"
	"backendSenior/utills"
	g "common/utils/ginutils"
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/globalsign/mgo"
	"os"
	"time"

	"backendSenior/domain/model"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/globalsign/mgo/bson"
)

type RoomRouteHandler struct {
	roomService  *service.RoomService
	userService  *service.UserService
	proxyService *service.ProxyService
	authMw       *auth.JWTMiddleware
	chatService  *service.ChatService
	orgService   *service.OrganizeService
	logger       *log.Logger
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
		logger:       log.New(os.Stdout, "RoomRouteHandler ", log.LstdFlags|log.Lshortfile),
	}
}

//Mount make RoomRouteHandler handler request from specific `RouterGroup`
func (handler *RoomRouteHandler) Mount(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/id/:id/member", handler.getRoomMember)
	routerGroup.POST("/id/:id/member", handler.authMw.AuthRequired(), g.InjectGin(handler.addMemberToRoom))
	routerGroup.DELETE("/id/:id/member", handler.deleteMemberFromRoom)
	// room admin API
	routerGroup.GET("/id/:id/admin", g.InjectGin(handler.getRoomAdmins))
	routerGroup.POST("/id/:id/admin", handler.authMw.AuthRequired(), g.InjectGin(handler.addAdminsToRoom))
	routerGroup.DELETE("/id/:id/admin", handler.authMw.AuthRequired(), g.InjectGin(handler.removeAdminsFromRoom))

	routerGroup.GET("/id/:id/proxy", handler.getRoomProxies)
	routerGroup.POST("/id/:id/proxy", handler.addProxiesToRoom)
	routerGroup.DELETE("/id/:id/proxy", handler.removeProxiesFromRoom)

	routerGroup.POST("/create-group", handler.authMw.AuthRequired(), g.InjectGin(handler.createGroupHandler))
	routerGroup.POST("/create-private-chat", handler.authMw.AuthRequired(), g.InjectGin(handler.createPrivateChatHandler))
	routerGroup.POST("/id/:id/name" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.editRoomNameHandler)
	routerGroup.DELETE("/id/:id" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.deleteRoomByIDHandler)
	// routerGroup.POST("/addmembertoroom" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.addMemberToRoom)
	// routerGroup.POST("/deletemembertoroom" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.deleteMemberFromRoom)
	routerGroup.GET("/id/:id" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.getRoomByIDHandler)
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

// for room group, we invite later
type createGroupDto struct {
	RoomName string        `validate:"required"`
	OrgId    bson.ObjectId `validate:"required"`
}

const (
	RoomTypeGroup   = "GROUP"
	RoomTypePrivate = "PRIVATE"
)

func (d *createGroupDto) toRoom() model.Room {
	return model.Room{
		RoomName: d.RoomName,
		// We dont have orgId here, since we want it to be set after room is "invited" to org
		CreatedTimeStamp: time.Now(),
		RoomType:         RoomTypeGroup,
		ListUser:         []bson.ObjectId{},
		ListProxy:        []bson.ObjectId{},
	}
}

// createGroupHandler create room with only you
// it's user responsibility to invite more user later
func (handler *RoomRouteHandler) createGroupHandler(c *gin.Context, input struct{ Body createGroupDto }) error {
	// TODO: should check org exists and is in org
	b := input.Body

	roomID, err := handler.roomService.AddRoom(b.toRoom())
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		return fmt.Errorf("error creating room: %s", err)
	}

	currentUserId := c.GetString(auth.UserIdField)
	if err := handler.roomService.AddMembersToRoom(roomID, []string{currentUserId}); err != nil {
		return fmt.Errorf("error adding self to room: %s", err)
	}

	if err := handler.orgService.AddRoomsToOrg(b.OrgId.Hex(), []string{roomID}); err != nil {
		return fmt.Errorf("error adding room to org: %s", err)
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "roomId": roomID})
	return nil
}

type createPrivateChatDto struct {
	RoomName string        `validate:"required"`
	OrgId    bson.ObjectId `validate:"required"`
	UserId   bson.ObjectId `validate:"required"`
}

func (d *createPrivateChatDto) toRoom() model.Room {
	return model.Room{
		RoomName:         d.RoomName,
		OrgID:            d.OrgId,
		CreatedTimeStamp: time.Now(),
		RoomType:         RoomTypePrivate,
		ListUser:         nil,
		ListProxy:        nil,
	}
}

// createPrivateChatHandler, create private chat with two person
// don't need to invite later
func (handler *RoomRouteHandler) createPrivateChatHandler(c *gin.Context, input struct {
	Body createPrivateChatDto
}) error {
	// TODO: ensure that...
	// - org exists
	// - both are in that org
	b := input.Body
	// check user in org
	currentUserId := c.GetString(auth.UserIdField)
	userInfo, err := handler.userService.GetUserByID(currentUserId)
	if err != nil {
		return err
	}
	// both user must be in the org
	allowed := false
	for _, org := range userInfo.Organize {
		if org == b.OrgId {
			allowed = true
			break
		}
	}
	if !allowed {
		return g.NewError(403, "you are not in specified org")
	}

	roomID, err := handler.roomService.AddRoom(b.toRoom())
	if err != nil {
		return err
	}
	if err := handler.roomService.AddMembersToRoom(roomID, []string{currentUserId, b.UserId.Hex()}); err != nil {
		return err
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"roomId": roomID,
	})
	return nil
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
func (handler *RoomRouteHandler) addMemberToRoom(c *gin.Context, input struct {
	Body struct {
		UserIDs []bson.ObjectId
	}
}) error {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "bad room ID")
	}

	body := input.Body
	joinIDs := utills.ToStringArr(body.UserIDs)
	err := handler.roomService.AddMembersToRoom(roomID, joinIDs)
	if err != nil {
		log.Println("error AddMemberToRoom", err.Error())
		return g.NewError(500, fmt.Sprintf("error adding self to room: %s", err))
	}

	// note: this method doesn't return error yet
	handler.chatService.BroadcastMessageToRoom(roomID, chatsocket.Message{
		Type: message_types.Room,
		Payload: room.MemberEvent{
			Type:    room.Join,
			RoomID:  roomID,
			Members: joinIDs,
		},
	})

	c.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
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

	leaveIDs := utills.ToStringArr(body.UserIDs)
	err = handler.roomService.DeleteMemberFromRoom(roomID, leaveIDs)

	if err != nil {
		log.Println("error DeleteRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	// note: this method doesn't return error yet
	handler.chatService.BroadcastMessageToRoom(roomID, chatsocket.Message{
		Type: message_types.Room,
		Payload: room.MemberEvent{
			Type:    room.Leave,
			RoomID:  roomID,
			Members: leaveIDs,
		},
	})

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

func (handler *RoomRouteHandler) uploadFile(c *gin.Context) {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "bad room id",
		}) // TODO
		return
	}

	fileHeader, err := c.FormFile("file")
	if fileHeader == nil || err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "bad room id",
		}) // TODO
		return
	}

	_, err = fileHeader.Open()
	if err != nil {
		log.Println("uploadFile:", err)
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "error reading file, try again",
		})
		return
	}

	// w := c.Writer
	// w.WriteHeader(200)

	// buf := make([]byte, 4<<10) // 4KB
	// for {
	// 	n, err := file.Read(buf)
	// 	if err != nil {
	// 		if err == io.EOF {
	// 			break
	// 		} else {
	// 			c.JSON(500, gin.H{
	// 				"status":  "error",
	// 				"message": "error reading file, try again",
	// 			})
	// 			return
	// 		}
	// 	}
	// 	w.Write(buf[])
	// }

}

type empty struct{}
type any = interface{}

type inviteAdminDto struct {
	UserIDs []bson.ObjectId `validate:"required,min=1,dive,required"`
}

// Room ADMIN API
// addAdminsToRoom add admins to room (auto invite them as member)
func (handler *RoomRouteHandler) addAdminsToRoom(c *gin.Context, req struct {
	Body inviteAdminDto
}) error {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "bad room ID in path")
	}

	room, err := handler.roomService.GetRoomByID(roomID)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, "room does not exist")
		}
		return g.NewError(500, "something went wrong")
	}

	if match := linq.From(room.ListAdmin).FirstWith(func(i any) bool {
		return i.(bson.ObjectId).Hex() == roomID
	}); match == nil {
		return g.NewError(403, "must be admin of selected room")
	}

	body := req.Body
	// dont allow adding outsider
	org, err := handler.orgService.GetOrganizeById(room.OrgID.Hex())
	if err != nil {
		return g.NewError(500, err.Error())
	}
	orgMemberSet := make(map[bson.ObjectId]bool, len(org.Members))
	linq.From(org.Members).ForEach(func(i any) {
		orgMemberSet[i.(bson.ObjectId)] = true
	})
	if allOk := linq.From(body.UserIDs).All(func(i any) bool {
		_, inOrg := orgMemberSet[i.(bson.ObjectId)]
		return inOrg
	}); !allOk {
		return g.NewError(400, "some member in list are not in org")
	}

	if err := handler.roomService.AddAdminsToRoom(room.RoomID, body.UserIDs); err != nil {
		handler.logger.Println("can't invite user to room", err)
		return g.NewError(500, err.Error())
	}

	c.JSON(200, g.Response{
		Success: true,
		Message: "invited admin",
	})
	return nil
}

func (handler *RoomRouteHandler) getRoomAdmins(c *gin.Context, req struct{ Body empty }) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad room id in path")
	}

	if room, err := handler.roomService.GetRoomByID(id); err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, "room not found")
		}
		return err
	} else {
		adminUsers, err := handler.userService.GetUsersByIDs(utills.ToStringArr(room.ListAdmin))
		if err != nil {
			handler.logger.Print("error getting users", err)
			return err
		}
		c.JSON(200, adminUsers)
		return nil
	}
}

// removeAdminsFromRoom demote admins to user, it doesn't kick them
// if want to kick, use deleteMemberFromRoom
func (handler *RoomRouteHandler) removeAdminsFromRoom(c *gin.Context, req struct {
	Body inviteAdminDto
}) error {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "bad room ID in path")
	}

	room, err := handler.roomService.GetRoomByID(roomID)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, "room does not exist")
		}
		return g.NewError(500, "something went wrong")
	}

	if match := linq.From(room.ListAdmin).FirstWith(func(i any) bool {
		return i.(bson.ObjectId).Hex() == roomID
	}); match == nil {
		return g.NewError(403, "must be admin of selected room")
	}

	user, err := handler.userService.GetUserByID(c.GetString(auth.UserIdField))
	if err != nil {
		return g.NewError(500, err.Error())
	}
	if match := linq.From(user.Organize).FirstWith(func(i any) bool {
		return i.(bson.ObjectId) == room.OrgID
	}); match == nil {
		return g.NewError(403, "foreign user can't invite")
	}

	body := req.Body
	if err := handler.roomService.RemoveAdminsFromRoom(room.RoomID, body.UserIDs); err != nil {
		handler.logger.Println("can't invite user to room", err)
		return g.NewError(500, err.Error())
	}
	c.JSON(200, g.Response{
		Success: true,
		Message: "invited admin",
	})
	return nil
}
