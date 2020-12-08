package routeAPI

import (
	repository "backendSenior/data/repository/mongo_repository"

	service "backendSenior/domain/usecase"
	"backendSenior/domain/usecase/auth"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

func AddRoomRoute(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	// prepare deps
	roomRepository := repository.RoomRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	roomService := service.NewRoomService(roomRepository)

	userRepository := repository.UserRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	authService := &auth.AuthService{
		UserRepository: &userRepository,
	}

	roomRouterHandler := NewRoomRouteHandler(roomService, authService)
	roomRouterHandler.Mount(routerGroup.Group("/v1"))
}

// func AddRoomRouteDev(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

// 	// Room API
// 	roomRepository := repository.RoomRepositoryMongo{
// 		ConnectionDB: connectionDB,
// 	}
// 	roomAPI := api.RoomAPI{
// 		RoomRepository: &roomRepository,
// 	}
// 	routerGroup.GET("/v1/room", roomAPI.RoomListHandler)
// 	routerGroup.POST("/v1/createroom", roomAPI.AddRoomHandeler)
// 	routerGroup.PUT("/v1/editroomname", roomAPI.EditRoomNameHandler)
// 	routerGroup.POST("/v1/deleteroom", roomAPI.DeleteRoomByIDHandler)
// 	routerGroup.POST("/v1/addmembertoroom", roomAPI.AddMemberToRoom)
// 	routerGroup.POST("/v1/deletemembertoroom", roomAPI.DeleteMemberToRoom)

// 	//Socket-API Call
// 	routerGroup.PUT("/v1/AddMemberToRoom", roomAPI.AddMemberToRoom)
// }