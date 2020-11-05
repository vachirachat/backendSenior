package routeAPI

import (
	"backendSenior/api"
	"backendSenior/api/auth"
	"backendSenior/repository"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

func AddRoomRoute(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	// Room API
	roomRepository := repository.RoomRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	roomAPI := api.RoomAPI{
		RoomRepository: &roomRepository,
	}

	userRepository := repository.UserRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	authAPI := auth.Auth{
		UserRepository: &userRepository,
	}
	routerGroup.GET("/v1/room", authAPI.AuthMiddleware("object", "view"), roomAPI.RoomListHandler)
	routerGroup.POST("/v1/room", authAPI.AuthMiddleware("object", "view"), roomAPI.AddRoomHandeler)
	routerGroup.PUT("/v1/room/:room_id", authAPI.AuthMiddleware("object", "view"), roomAPI.EditRoomNameHandler)
	routerGroup.DELETE("/v1/room", authAPI.AuthMiddleware("object", "view"), roomAPI.DeleteRoomByIDHandler)

}

func AddRoomRouteDev(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	// Room API
	roomRepository := repository.RoomRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	roomAPI := api.RoomAPI{
		RoomRepository: &roomRepository,
	}
	routerGroup.GET("/v1/room", roomAPI.RoomListHandler)
	routerGroup.POST("/v1/room", roomAPI.AddRoomHandeler)
	routerGroup.PUT("/v1/room/:room_id", roomAPI.EditRoomNameHandler)
	routerGroup.DELETE("/v1/room", roomAPI.DeleteRoomByIDHandler)

	//Socket-API Call
	routerGroup.PUT("/v1/invitePeopleRoom", roomAPI.AddMemberToRoom)
}
