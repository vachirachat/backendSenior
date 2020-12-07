package routeAPI

import (
	repository "backendSenior/data/repository/mongo_repository"

	api "backendSenior/domain/usecase"
	"backendSenior/domain/usecase/auth"

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
	// routerGroup.GET("/v1/room", authAPI.AuthMiddleware("object", "view"), roomAPI.RoomListHandler)
	routerGroup.POST("/v1/createroom", authAPI.AuthMiddleware("object", "view"), roomAPI.AddRoomHandeler)
	routerGroup.PUT("/v1/editroomname", authAPI.AuthMiddleware("object", "view"), roomAPI.EditRoomNameHandler)
	routerGroup.DELETE("/v1/deleteroom", authAPI.AuthMiddleware("object", "view"), roomAPI.DeleteRoomByIDHandler)
	routerGroup.POST("/v1/addmembertoroom", authAPI.AuthMiddleware("object", "view"), roomAPI.AddMemberToRoom)
	routerGroup.POST("/v1/deletemembertoroom", authAPI.AuthMiddleware("object", "view"), roomAPI.DeleteMemberToRoom)
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
	routerGroup.POST("/v1/createroom", roomAPI.AddRoomHandeler)
	routerGroup.PUT("/v1/editroomname", roomAPI.EditRoomNameHandler)
	routerGroup.POST("/v1/deleteroom", roomAPI.DeleteRoomByIDHandler)
	routerGroup.POST("/v1/addmembertoroom", roomAPI.AddMemberToRoom)
	routerGroup.POST("/v1/deletemembertoroom", roomAPI.DeleteMemberToRoom)

	//Socket-API Call
	routerGroup.PUT("/v1/AddMemberToRoom", roomAPI.AddMemberToRoom)
}
