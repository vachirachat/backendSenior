package routeAPI

import (
	repository "backendSenior/data/repository/mongo_repository"

	service "backendSenior/domain/usecase"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

func AddUserRoute(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	userRepository := repository.UserRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	userService := service.NewUserService(userRepository)

	userRouteHandler := NewUserRouteHandler(userService)

	userRouteHandler.Mount(routerGroup.Group("/v1"))

	// authService := auth.AuthService{
	// 	UserRepository: &userRepository,
	// }

}

// func AddUserRouteDev(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

// 	userRepository := repository.UserRepositoryMongo{
// 		ConnectionDB: connectionDB,
// 	}
// 	userAPI := service.UserAPI{
// 		UserRepository: &userRepository,
// 	}

// 	routerGroup.GET("/v1/user", userAPI.UserListHandler)
// 	routerGroup.POST("/v1/user", userAPI.AddUserHandeler)
// 	//routerGroup.POST("/v1/getroomuser", userAPI.GetUserRoomByUserID)
// 	routerGroup.POST("/v1/updateuser", userAPI.UpdateUserHandler)
// 	routerGroup.PUT("/v1/updateuserprofile", userAPI.EditUserNameHandler)
// 	routerGroup.DELETE("/v1/user/:user_id", userAPI.DeleteUserByIDHandler)
// 	routerGroup.POST("/v1/getuserbyemail", userAPI.GetUserByEmail)
// 	routerGroup.GET("/v1/getlistSecret", userAPI.GetUserListSecrect)
// }
