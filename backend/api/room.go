package api

import (
	"backendSenior/model"
	"backendSenior/repository"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoomAPI struct {
	RoomRepository repository.RoomRepository
}

func (api RoomAPI) RoomListHandler(context *gin.Context) {
	var roomsInfo model.RoomInfo
	rooms, err := api.RoomRepository.GetAllRoom()
	if err != nil {
		log.Println("error roomListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	roomsInfo.Room = rooms
	context.JSON(http.StatusOK, roomsInfo)
}

// for get room by id
func (api RoomAPI) GetRoomByIDHandler(context *gin.Context) {
	roomID := context.Param("roomId")
	room, err := api.RoomRepository.GetRoomByID(roomID)
	if err != nil {
		log.Println("error GetRoomByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusOK, room)
}

func (api RoomAPI) AddRoomHandeler(context *gin.Context) {
	var room model.Room
	err := context.ShouldBindJSON(&room)
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	err = api.RoomRepository.AddRoom(room)
	if err != nil {
		log.Println("error AddRoomHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (api RoomAPI) EditRoomNameHandler(context *gin.Context) {
	var room model.Room
	roomID := context.Param("room_id")
	err := context.ShouldBindJSON(&room)
	if err != nil {
		log.Println("error EditProducNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	err = api.RoomRepository.EditRoomName(roomID, room)
	if err != nil {
		log.Println("error EditProducNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	err = api.RoomRepository.DeleteRoomByID(room.RoomID)
	if err != nil {
		log.Println("error DeleteRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
	context.JSON(http.StatusNoContent, gin.H{"message": "success"})
}

// Match with Socket-structure

//// -- JoinAPI -> getSession(Topic+#ID) -> giveUserSession
func (api RoomAPI) InviteUserByIDHandler(context *gin.Context) {
	//roomID := context.Param("room_id")
	var room model.Room
	err := context.ShouldBindJSON(&room)

	if err != nil {
		log.Println("error DeleteRoomByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// err = api.RoomRepository.DeleteRoomByID(room.RoomID)
	// if err != nil {
	// 	log.Println("error DeleteRoomHandler", err.Error())
	// 	context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	// }
	context.JSON(http.StatusNoContent, gin.H{"message": "success"})
}
