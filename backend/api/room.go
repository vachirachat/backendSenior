package api

import (
	"backendSenior/model"
	"backendSenior/repository"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

type RoomAPI struct {
	RoomRepository repository.RoomRepository
}

func (api RoomAPI) RoomListHandler(context *gin.Context) {
	var roomsInfo model.RoomInfo
	rooms, err := api.RoomRepository.GetAllRoom()
	if err != nil {
		log.Println("error roomListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	roomsInfo.Room = rooms
	context.JSON(http.StatusOK, roomsInfo)
}

// for get room by id
func (api RoomAPI) GetRoomByIDHandler(context *gin.Context) {
	roomID := context.Param("roomId")
	ObjectID := bson.ObjectIdHex(roomID)
	room, err := api.RoomRepository.GetRoomByID(ObjectID)
	if err != nil {
		log.Println("error GetRoomByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, room)
}

func (api RoomAPI) AddRoomHandeler(context *gin.Context) {
	var room model.Room
	err := context.ShouldBindJSON(&room)
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	err = api.RoomRepository.AddRoom(room)
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (api RoomAPI) EditRoomNameHandler(context *gin.Context) {
	var room model.Room
	err := context.ShouldBindJSON(&room)
	if err != nil {
		log.Println("error EditRoomNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	err = api.RoomRepository.EditRoomName(room.RoomID, room)
	if err != nil {
		log.Println("error EditRoomNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (api RoomAPI) DeleteRoomByIDHandler(context *gin.Context) {
	//roomID := context.Param("room_id")
	var room model.Room
	err := context.ShouldBindJSON(&room)
	log.Println(room)
	if err != nil {
		log.Println("error DeleteRoomByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	err = api.RoomRepository.DeleteRoomByID(room.RoomID)
	if err != nil {
		log.Println("error DeleteRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Match with Socket-structure

//// -- JoinAPI -> getSession(Topic+#ID) -> giveUserSession
func (api RoomAPI) AddMemberToRoom(context *gin.Context) {
	// send JSON Body
	/* - {
		"roomID" : ""
		"ListUser" : [""]
	}*/

	var room model.Room
	err := context.ShouldBindJSON(&room)
	if err != nil {
		log.Println("error InviteUserByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	err = api.RoomRepository.AddMemberToRoom(room.RoomID, room.ListUser)
	if err != nil {
		log.Println("error AddMemberToRoom", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (api RoomAPI) DeleteMemberToRoom(context *gin.Context) {
	type deleteRoom struct {
		userID bson.ObjectId
		roomID bson.ObjectId
	}
	var roomDelete deleteRoom
	err := context.ShouldBindJSON(&roomDelete)
	if err != nil {
		log.Println("error InviteUserByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	err = api.RoomRepository.DeleteMemberToRoom(roomDelete.userID, roomDelete.roomID)
	if err != nil {
		log.Println("error DeleteRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}
