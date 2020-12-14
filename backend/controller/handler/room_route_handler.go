package route

import (
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"

	"backendSenior/domain/model"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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
	routerGroup.POST("/createroom" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.addRoomHandler)
	routerGroup.PUT("/editroomname" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.editRoomNameHandler)
	routerGroup.DELETE("/deleteroom" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.deleteRoomByIDHandler)
	routerGroup.POST("/addmembertoroom" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.addMemberToRoom)
	routerGroup.POST("/deletememberfromroom" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.deleteMemberFromRoom)
	routerGroup.POST("/getroombyid" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.getRoomByIDHandler)
	routerGroup.GET("/listroom" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.roomListHandler)
	routerGroup.GET("/getroommember" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.getRoomMemberHandler)
}

func (handler *RoomRouteHandler) getRoomMemberHandler(context *gin.Context) {
	var body struct {
		RoomID string `json:"roomId"`
	}
	err := context.ShouldBindJSON(&body)
	rooms, err := handler.roomService.GetRoomByID(body.RoomID)
	if err != nil {
		log.Println("error getRoomMemberHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success", "listUser": rooms.ListUser})
}

func (handler *RoomRouteHandler) roomListHandler(context *gin.Context) {
	var roomsInfo model.RoomInfo
	rooms, err := handler.roomService.GetAllRooms()
	if err != nil {
		log.Println("error roomListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	roomsInfo.Room = model.ArrRoomMongoToRoomString(rooms)
	context.JSON(http.StatusOK, roomsInfo)
}

// for get room by id
func (handler *RoomRouteHandler) getRoomByIDHandler(context *gin.Context) {
	var room model.Room
	err := context.ShouldBindJSON(&room)
	if err != nil {
		log.Println("error GetRoomByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	room, err = handler.roomService.GetRoomByID(room.RoomID)
	if err != nil {
		log.Println("error GetRoomByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	context.JSON(http.StatusOK, room.RoomStringIDToMongoID())
}

func (handler *RoomRouteHandler) addRoomHandler(context *gin.Context) {
	var room model.Room
	err := context.ShouldBindJSON(&room)
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
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
	var room model.Room
	err := context.ShouldBindJSON(&room)
	if err != nil {
		log.Println("error EditRoomNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	log.Println(room)
	err = handler.roomService.EditRoomName(room.RoomID, room)
	if err != nil {
		log.Println("error EditRoomNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *RoomRouteHandler) deleteRoomByIDHandler(context *gin.Context) {
	//roomID := context.Param("room_id")
	var room model.Room
	err := context.ShouldBindJSON(&room)
	if err != nil {
		log.Println("error DeleteRoomByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	err = handler.roomService.DeleteRoomByID(room.RoomID)
	if err != nil {
		log.Println("error DeleteRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Match with Socket-structure

//// -- JoinAPI -> getSession(Topic+#ID) -> giveUserSession
func (handler *RoomRouteHandler) addMemberToRoom(context *gin.Context) {
	var body struct {
		RoomID   string   `json:"roomId" bson:"roomId"`
		UserList []string `json:"userIds" bson:"userIds"`
	}

	err := context.ShouldBindJSON(&body)
	log.Println(body)
	if err != nil {
		log.Println("error InviteUserByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	err = handler.roomService.AddMembersToRoom(body.RoomID, body.UserList)
	if err != nil {
		log.Println("error AddMemberToRoom", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *RoomRouteHandler) deleteMemberFromRoom(context *gin.Context) {
	var body struct {
		RoomID   string   `json:"roomId" bson:"roomId"`
		UserList []string `json:"userIds" bson:"userIds"`
	}

	err := context.ShouldBindJSON(&body)
	if err != nil {
		log.Println("error InviteUserByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	err = handler.roomService.DeleteMemberFromRoom(body.RoomID, body.UserList)

	if err != nil {
		log.Println("error DeleteRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}
