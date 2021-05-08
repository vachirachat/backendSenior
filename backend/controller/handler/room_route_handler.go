package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/dto"
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/message_types"
	"backendSenior/domain/model/chatsocket/room"
	"backendSenior/domain/service"
	"backendSenior/utills"
	g "common/utils/ginutils"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/ahmetb/go-linq/v3"
	"github.com/globalsign/mgo"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

type RoomRouteHandler struct {
	roomService  *service.RoomService
	userService  *service.UserService
	proxyService *service.ProxyService
	authMw       *auth.JWTMiddleware
	proxyMw      *auth.ProxyMiddleware
	chatService  *service.ChatService
	orgService   *service.OrganizeService
	logger       *log.Logger
	keyService   *service.KeyExchangeService
	fileService  *service.FileService
	validate     *utills.StructValidator
}

// NewRoomHandler create new handler for room
func NewRoomRouteHandler(roomService *service.RoomService,
	authMw *auth.JWTMiddleware,
	proxyMw *auth.ProxyMiddleware,
	userService *service.UserService,
	proxyService *service.ProxyService,
	chatService *service.ChatService,
	orgService *service.OrganizeService,
	keyService *service.KeyExchangeService,
	fileService *service.FileService,
	validate *utills.StructValidator,
) *RoomRouteHandler {
	return &RoomRouteHandler{
		roomService:  roomService,
		userService:  userService,
		authMw:       authMw,
		proxyMw:      proxyMw,
		proxyService: proxyService,
		chatService:  chatService,
		orgService:   orgService,
		keyService:   keyService,
		logger:       log.New(os.Stdout, "RoomRouteHandler ", log.LstdFlags|log.Lshortfile),
		fileService:  fileService,
		validate:     validate,
	}
}

//Mount make RoomRouteHandler handler request from specific `RouterGroup`
func (handler *RoomRouteHandler) Mount(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/id/:id/member", handler.proxyMw.AlternativeAuth(), handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInRoomMiddleWare("id"), g.InjectGin(handler.getRoomMember))
	routerGroup.POST("/id/:id/member", handler.authMw.AuthRequired("user", "add"), handler.authMw.IsInRoomMiddleWare("id"), g.InjectGin(handler.addMemberToRoom))
	routerGroup.DELETE("/id/:id/member", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsInRoomMiddleWare("id"), g.InjectGin(handler.deleteMemberFromRoom))
	// room admin API
	routerGroup.GET("/id/:id/admin", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInRoomMiddleWare("id"), g.InjectGin(handler.getRoomAdmins))
	routerGroup.POST("/id/:id/admin", handler.authMw.AuthRequired("user", "add"), handler.authMw.IsRoomAdmitMiddleWare("id"), g.InjectGin(handler.addAdminsToRoom))
	routerGroup.DELETE("/id/:id/admin", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsRoomAdmitMiddleWare("id"), g.InjectGin(handler.removeAdminsFromRoom))

	routerGroup.GET("/id/:id/proxy", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsRoomAdmitMiddleWare("id"), g.InjectGin(handler.getRoomProxies))
	routerGroup.POST("/id/:id/proxy", handler.authMw.AuthRequired("user", "add"), handler.authMw.IsRoomAdmitMiddleWare("id"), g.InjectGin(handler.addProxiesToRoom))
	routerGroup.DELETE("/id/:id/proxy", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsRoomAdmitMiddleWare("id"), g.InjectGin(handler.removeProxiesFromRoom))

	routerGroup.POST("/id/:id/image", handler.authMw.IsInRoomMiddleWare("id"), g.InjectGin(handler.uploadRoomImage))
	routerGroup.GET("/id/:id/image", handler.authMw.IsInRoomMiddleWare("id"), g.InjectGin(handler.getRoomImage))

	routerGroup.POST("/create-group", handler.authMw.AuthRequired("user", "view"), g.InjectGin(handler.createGroupHandler))
	routerGroup.POST("/create-private-chat", handler.authMw.AuthRequired("user", "view"), g.InjectGin(handler.createPrivateChatHandler))
	routerGroup.POST("/id/:id/name", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInRoomMiddleWare("id"), g.InjectGin(handler.editRoomNameHandler))
	routerGroup.DELETE("/id/:id", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsRoomAdmitMiddleWare("id"), g.InjectGin(handler.deleteRoomByIDHandler))

	routerGroup.GET("/id/:id", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInRoomMiddleWare("id"), g.InjectGin(handler.getRoomByIDHandler))
	routerGroup.GET("/", handler.authMw.AuthRequired("user", "view"), g.InjectGin(handler.roomListHandler))
}

func (handler *RoomRouteHandler) roomListHandler(context *gin.Context, req struct{}) error {
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
		return g.NewError(404, "bad Get listRoom")
	}
	roomsInfo.Room = rooms
	context.JSON(http.StatusOK, roomsInfo)
	return nil
}

// for get room by id
func (handler *RoomRouteHandler) getRoomByIDHandler(context *gin.Context, req struct{}) error {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "room id in path")
	}

	room, err := handler.roomService.GetRoomByID(roomID)
	if err != nil {
		return g.NewError(404, "bad get room")
	}
	context.JSON(http.StatusOK, room)
	return nil
}

// // for room group, we invite later
// type createGroupDto struct {
// 	RoomName string        `validate:"required"`
// 	OrgId    bson.ObjectId `validate:"required"`
// }

// func (d *createGroupDto) toRoom() model.Room {
// 	return model.Room{
// 		RoomName: d.RoomName,
// 		// We dont have orgId here, since we want it to be set after room is "invited" to org
// 		CreatedTimeStamp: time.Now(),
// 		RoomType:         model.RoomGroup,
// 		ListUser:         []bson.ObjectId{},
// 		ListProxy:        []bson.ObjectId{},
// 	}
// }

// createGroupHandler create room with only you
// it's user responsibility to invite more user later
func (handler *RoomRouteHandler) createGroupHandler(c *gin.Context, input struct{ Body dto.CreateGroupDto }) error {
	// TODO: should check org exists and is in org
	b := input.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}
	roomID, err := handler.roomService.AddRoom(b.ToRoom())
	if err != nil {
		return g.NewError(403, "bad Add Room")
	}
	// validate org
	_, err = handler.orgService.GetOrganizeById(b.OrgId.Hex())
	if err != nil {
		return g.NewError(403, "bad get orgID")
	}

	currentUserID := c.GetString(auth.UserIdField)
	if err := handler.roomService.AddAdminsToRoom(bson.ObjectIdHex(roomID), []bson.ObjectId{bson.ObjectIdHex(currentUserID)}); err != nil {
		return g.NewError(403, "bad  adding self to room")
	}

	if err := handler.orgService.AddRoomsToOrg(b.OrgId.Hex(), []string{roomID}); err != nil {
		return g.NewError(403, "bad  adding room to org")
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "roomId": roomID})
	return nil
}

// createPrivateChatHandler, create private chat with two person
// don't need to invite later
func (handler *RoomRouteHandler) createPrivateChatHandler(c *gin.Context, input struct {
	Body dto.CreatePrivateChatDto
}) error {
	// TODO: ensure that...
	// - org exists
	// - both are in that org
	b := input.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}
	// check user in org
	currentUserID := c.GetString(auth.UserIdField)
	userInfo, err := handler.userService.GetUserByID(currentUserID)
	if err != nil {
		return g.NewError(403, "bad get UserId")
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

	roomID, err := handler.roomService.AddRoom(b.ToRoom())
	if err != nil {
		return err
	}
	if err := handler.roomService.AddMembersToRoom(roomID, []string{currentUserID, b.UserId.Hex()}); err != nil {
		return err
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"roomId": roomID,
	})
	return nil
}

func (handler *RoomRouteHandler) editRoomNameHandler(context *gin.Context, req struct {
	Body struct {
		RoomName string `json:"roomName" validate:"required,gt=0"`
	}
}) error {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "room id in path")
	}

	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	err = handler.roomService.EditRoomName(roomID, model.Room{RoomName: b.RoomName})
	if err != nil {
		return err
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

func (handler *RoomRouteHandler) deleteRoomByIDHandler(context *gin.Context, req struct{}) error {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "room id in path")
	}

	err := handler.roomService.DeleteRoomByID(roomID)
	if err != nil {
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

// Match with Socket-structure

//// -- JoinAPI -> getSession(Topic+#ID) -> giveUserSession
func (handler *RoomRouteHandler) addMemberToRoom(c *gin.Context, req struct {
	Body struct {
		UserIDs []bson.ObjectId `json:"userIDs" validate:"required,dive,gt=0"`
	}
}) error {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "bad room ID")
	}
	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	// use in room
	currentRoom, err := handler.roomService.GetRoomByID(roomID)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, "room not found")
		}
		return err
	}
	if currentRoom.RoomType != model.RoomGroup {
		return g.NewError(400, "not allowed on private room")
	}

	userID := bson.ObjectIdHex(c.GetString(auth.UserIdField))
	if !linq.From(currentRoom.ListUser).Contains(userID) {
		return g.NewError(403, "you are not in the room")
	}
	// user in room's org
	org, err := handler.orgService.GetOrganizeById(currentRoom.OrgID.Hex())
	if err != nil {
		return err
	}
	if !linq.From(org.Members).Contains(userID) {
		return g.NewError(403, "cross-org user not allowed to invite")
	}

	// check invited member are not cross org
	if !linq.From(org.Admins).Contains(userID) {
		handler.logger.Println("check invited member in org")
		orgMemberSet := make(map[bson.ObjectId]bool, len(org.Members))
		linq.From(org.Members).ForEach(func(i dto.Any) {
			orgMemberSet[i.(bson.ObjectId)] = true
		})
		if allOk := linq.From(req.Body.UserIDs).All(func(i dto.Any) bool {
			_, inOrg := orgMemberSet[i.(bson.ObjectId)]
			return inOrg
		}); !allOk {
			return g.NewError(400, "as room member, not allowed to invite cross-org")
		}
	}

	joinIDs := utills.ToStringArr(req.Body.UserIDs)
	if err := handler.roomService.AddMembersToRoom(roomID, joinIDs); err != nil {
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

func (handler *RoomRouteHandler) deleteMemberFromRoom(c *gin.Context, req struct {
	Body struct {
		UserIDs []bson.ObjectId `json:"userIDs" validate:"required,min=1,dive,required"` //TODO
	}
}) error {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "bad param roomID")
	}
	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	currentRoom, err := handler.roomService.GetRoomByID(roomID)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, "room not found")
		}
		return err
	}
	if currentRoom.RoomType != model.RoomGroup {
		return g.NewError(400, "not allowed on private room")
	}

	userID := bson.ObjectIdHex(c.GetString(auth.UserIdField))
	if !linq.From(currentRoom.ListUser).Contains(userID) {
		return g.NewError(403, "you are not in the room")
	}
	// user in room's org
	org, err := handler.orgService.GetOrganizeById(currentRoom.OrgID.Hex())
	if err != nil {
		return err
	}
	if !linq.From(org.Members).Contains(userID) {
		return g.NewError(403, "cross-org user not allowed to kick")
	}

	// if not admin, the allow only kick non-admin
	if !linq.From(currentRoom.ListAdmin).Contains(userID) {
		handler.logger.Println("check kicked member not admin")
		roomAdminSet := make(map[bson.ObjectId]bool, len(currentRoom.ListAdmin))
		linq.From(currentRoom.ListAdmin).ForEach(func(i dto.Any) {
			roomAdminSet[i.(bson.ObjectId)] = true
		})
		fmt.Printf("users %+v\n", req.Body.UserIDs)
		if allOk := linq.From(req.Body.UserIDs).All(func(i dto.Any) bool {
			// must not be admin
			_, isAdmin := roomAdminSet[i.(bson.ObjectId)]
			fmt.Printf("ID %v is admin ? %v\n", i, isAdmin)
			return !isAdmin
		}); !allOk {
			return g.NewError(400, "as member of room, not allowed to kick admin")
		}
	}

	leaveIDs := utills.ToStringArr(req.Body.UserIDs)
	if err := handler.roomService.DeleteMemberFromRoom(roomID, leaveIDs); err != nil {
		handler.logger.Println("error DeleteRoomHandler", err)
		return err
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

	c.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

// TODO change this to return full object only, currently keep for compatibility of proxy
func (handler *RoomRouteHandler) getRoomMember(context *gin.Context, req struct{}) error {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "room id in path")
	}
	isFull := context.Query("full") != ""

	userIDs, err := handler.roomService.GetRoomMemberIDs(roomID)
	if err != nil {
		return err
	}

	if !isFull {
		context.JSON(http.StatusOK, gin.H{
			"members": userIDs,
		})
		return nil
	}
	users, err := handler.userService.GetUsersByIDs(userIDs)
	if err != nil {
		return g.NewError(404, "bad getRoomMember")
	}

	context.JSON(http.StatusOK, gin.H{
		"members": users,
	})
	return nil
}

func (handler *RoomRouteHandler) getRoomProxies(context *gin.Context, req struct{}) error {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "room id in path")
	}

	proxyIDs, err := handler.roomService.GetRoomProxyIDs(roomID)
	if err != nil {
		return g.NewError(404, "bad getRoomProxies")
	}

	proxies, err := handler.proxyService.GetProxiesByIDs(proxyIDs)
	if err != nil {
		return g.NewError(403, "bad GetProxiesIDs")
	}

	context.JSON(http.StatusOK, gin.H{
		"proxies": proxies,
	})
	return nil

}

func (handler *RoomRouteHandler) addProxiesToRoom(context *gin.Context, req struct {
	Body struct {
		ProxyIDs []bson.ObjectId `json:"proxyIds" validate:"required,min=1,dive,required"`
	}
}) error {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "room id in path")
	}

	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	if err := handler.roomService.AddProxiesToRoom(roomID, utills.ToStringArr(b.ProxyIDs)); err != nil {
		return g.NewError(403, "bad  addProxiesToRoom: add proxy")
	}

	// Fix : find roomOrg and update roomOrg to proxy
	room, err := handler.roomService.GetRoomByID(roomID)
	if err != nil {
		return g.NewError(404, "bad  GetRoomByID")
	}

	if err := handler.proxyService.AddProxiseToOrg(room.OrgID.Hex(), utills.ToStringArr(b.ProxyIDs)); err != nil {
		return g.NewError(403, "bad  AddProxiseToOrg")
	}
	// Fix : find roomOrg and update roomOrg to proxy

	if err := handler.keyService.Ensure(roomID, utills.ToStringArr(b.ProxyIDs)); err != nil {
		handler.logger.Println("[Room handler] addProxiesToRoom: ensure key ", err.Error())
		return g.NewError(403, "bad addProxiesToRoom: ensure key ")
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil

}

func (handler *RoomRouteHandler) removeProxiesFromRoom(context *gin.Context, req struct {
	Body struct {
		ProxyIDs []bson.ObjectId `json:"proxyIds" validate:"required,min=1,dive,required"`
	}
}) error {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "room id in path")
	}

	b := req.Body

	err := handler.roomService.DeleteProxiesFromRoom(roomID, utills.ToStringArr(b.ProxyIDs))
	if err != nil {
		handler.logger.Println("[Room handler] removeProxiesFromRoom: ", err.Error())
		return err
	}
	if err := handler.keyService.Delete(roomID, utills.ToStringArr(b.ProxyIDs)); err != nil {
		handler.logger.Println("[Room handler] removeProxiesFromRoom: remove key ", err.Error())
		return err
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

func (handler *RoomRouteHandler) uploadFile(c *gin.Context, req struct{}) error {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "bad room id",
		}) // TODO
		return fmt.Errorf("error upload File")
	}

	fileHeader, err := c.FormFile("file")
	if fileHeader == nil || err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "bad room id",
		}) // TODO
		return err
	}

	_, err = fileHeader.Open()
	if err != nil {
		log.Println("uploadFile:", err)
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "error reading file, try again",
		})
		return err
	}
	return nil
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

// type empty struct{}
// type any = interface{}

// type InviteAdminDto struct {
// 	UserIDs []bson.ObjectId `validate:"required,min=1,dive,required"`
// }

// Room ADMIN API
// addAdminsToRoom add admins to room (auto invite them as member)
func (handler *RoomRouteHandler) addAdminsToRoom(c *gin.Context, req struct {
	Body dto.InviteAdminDto
}) error {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "bad currentRoom ID in path")
	}

	currentRoom, err := handler.roomService.GetRoomByID(roomID)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, "currentRoom does not exist")
		}
		return err
	}

	if currentRoom.RoomType != model.RoomGroup {
		return g.NewError(400, "not allowed on private room")
	}

	// NOTE: we are managing admin, so the user who request must be admin
	// admin can't be cross-org user, so no need extra check
	if !linq.From(currentRoom.ListAdmin).Contains(bson.ObjectIdHex(c.GetString(auth.UserIdField))) {
		return g.NewError(403, "must be admin of selected currentRoom")
	}

	body := req.Body
	// dont allow adding outsider as admin
	org, err := handler.orgService.GetOrganizeById(currentRoom.OrgID.Hex())
	if err != nil {
		return g.NewError(500, err.Error())
	}
	orgMemberSet := make(map[bson.ObjectId]bool, len(org.Members))
	linq.From(org.Members).ForEach(func(i dto.Any) {
		orgMemberSet[i.(bson.ObjectId)] = true
	})
	if allOk := linq.From(body.UserIDs).All(func(i dto.Any) bool {
		_, inOrg := orgMemberSet[i.(bson.ObjectId)]
		return inOrg
	}); !allOk {
		return g.NewError(400, "not allowed to invite cross-org user as admin")
	}

	if err := handler.roomService.AddAdminsToRoom(currentRoom.RoomID, body.UserIDs); err != nil {
		handler.logger.Println("can't invite user to currentRoom", err)
		return g.NewError(500, err.Error())
	}

	// note: this method doesn't return error yet
	handler.chatService.BroadcastMessageToRoom(roomID, chatsocket.Message{
		Type: message_types.Room,
		Payload: room.MemberEvent{
			Type:    room.Join,
			RoomID:  roomID,
			Members: utills.ToStringArr(body.UserIDs),
		},
	})

	c.JSON(200, g.Response{
		Success: true,
		Message: "invited admin",
	})
	return nil
}

func (handler *RoomRouteHandler) getRoomAdmins(c *gin.Context, req struct{}) error {
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
	Body dto.InviteAdminDto
}) error {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "bad currentRoom ID in path")
	}

	currentRoom, err := handler.roomService.GetRoomByID(roomID)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, "currentRoom does not exist")
		}
		return g.NewError(500, "something went wrong")
	}
	if currentRoom.RoomType != model.RoomGroup {
		return g.NewError(400, "not allowed on private room")
	}

	if !linq.From(currentRoom.ListAdmin).Contains(bson.ObjectIdHex(c.GetString(auth.UserIdField))) {
		return g.NewError(403, "must be admin of selected currentRoom")
	}

	body := req.Body
	if err := handler.roomService.RemoveAdminsFromRoom(currentRoom.RoomID, body.UserIDs); err != nil {
		handler.logger.Println("can't invite user to currentRoom", err)
		return g.NewError(500, err.Error())
	}
	c.JSON(200, g.Response{
		Success: true,
		Message: "removed admin",
	})
	return nil
}

func (handler *RoomRouteHandler) uploadRoomImage(c *gin.Context, req struct{}) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad room id")
	}

	if _, err := handler.roomService.GetRoomByID(id); err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, "room not found")
		}
		return err
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return g.NewError(400, fmt.Sprintf("can't get image form field: %s", err))
	}
	file, err := fileHeader.Open()
	if err != nil {
		return g.NewError(400, fmt.Sprintf("can't open file: %s", err))
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return g.NewError(400, fmt.Sprintf("can't read file: %s", err))
	}

	if err := handler.fileService.UploadRoomImage(id, data); err != nil {
		return err
	}
	c.JSON(200, g.OK("uploaded room image"))
	return nil
}

func (handler *RoomRouteHandler) getRoomImage(c *gin.Context, req struct{}) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad room id")
	}

	isThumb := c.Query("thumbnail") == "true"

	img, err := handler.fileService.GetRoomImage(id, isThumb)
	if err != nil {
		return err
	}

	c.Header("Content-Disposition", "inline")
	c.Header("Content-Length", fmt.Sprint(len(img)))
	c.Data(200, "image/jpeg", img)
	return nil
}
