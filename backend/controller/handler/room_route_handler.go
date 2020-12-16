package route

import (
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"
	"backendSenior/utills"
	"fmt"

	"backendSenior/domain/model"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/globalsign/mgo/bson"
)

type RoomRouteHandler struct {
	roomService *service.RoomService
	authService *auth.AuthService
}

// NewRoomHandler create new handler for room
func NewRoomRouteHandler(roomService *service.RoomService, authService *auth.AuthService) *RoomRouteHandler {
	return &RoomRouteHandler{
		roomService: roomService,
		authService: authService,
	}
}

//Mount make RoomRouteHandler handler request from specific `RouterGroup`
func (handler *RoomRouteHandler) Mount(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/:id/member", handler.getRoomMemberIDs)
	routerGroup.GET("/:id/proxy", handler.getRoomProxies)
	routerGroup.POST("/:id/proxy", handler.addProxiesToRoom)
	routerGroup.DELETE("/:id/proxy", handler.removeProxiesFromRoom)

	routerGroup.POST("/" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.addRoomHandler)
	routerGroup.POST("/:id/name" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.editRoomNameHandler)
	routerGroup.DELETE("/:id" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.deleteRoomByIDHandler)
	// routerGroup.POST("/addmembertoroom" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.addMemberToRoom)
	// routerGroup.POST("/deletemembertoroom" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.deleteMemberFromRoom)
	routerGroup.GET("/:id" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.getRoomByIDHandler)
	routerGroup.GET("/" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.roomListHandler)
}

func (handler *RoomRouteHandler) roomListHandler(context *gin.Context) {
	var roomsInfo model.RoomInfo
	rooms, err := handler.roomService.GetAllRooms()
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

func (handler *RoomRouteHandler) addRoomHandler(context *gin.Context) {
	var room model.Room
	err := context.ShouldBindJSON(&room)
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	roomID, err := handler.roomService.AddRoom(room)
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
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

	err = handler.roomService.AddMembersToRoom(roomID, utills.ToStringArr(body.UserIDs))
	if err != nil {
		log.Println("error AddMemberToRoom", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
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

	err = handler.roomService.DeleteMemberFromRoom(roomID, utills.ToStringArr(body.UserIDs))

	if err != nil {
		log.Println("error DeleteRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *RoomRouteHandler) getRoomMemberIDs(context *gin.Context) {
	roomID := context.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad roomID"})
		return
	}

	users, err := handler.roomService.GetRoomMembers(roomID)
	if err != nil {
		fmt.Println("[Room handler] getRoomMembers: ", err.Error())
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

	proxies, err := handler.roomService.GetRoomProxies(roomID)
	if err != nil {
		fmt.Println("[Room handler] getRoomMembers: ", err.Error())
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
