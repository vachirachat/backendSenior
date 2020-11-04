package routeAPI

import (
	"backendSenior/api"
	"backendSenior/api/auth"
	"backendSenior/repository"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

func AddUserRoute(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	userRepository := repository.UserRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	userAPI := api.UserAPI{
		UserRepository: &userRepository,
	}
	authAPI := auth.Auth{
		UserRepository: &userRepository,
	}
	routerGroup.GET("/v1/user", authAPI.AuthMiddleware("object", "view"), userAPI.UserListHandler)
	routerGroup.POST("/v1/user", authAPI.AuthMiddleware("object", "view"), userAPI.AddUserHandeler)
	routerGroup.PUT("/v1/user/updateuserprofile", authAPI.AuthMiddleware("object", "view"), userAPI.EditUserNameHandler)
	routerGroup.DELETE("/v1/user/:user_id", authAPI.AuthMiddleware("object", "view"), userAPI.DeleteUserByIDHandler)
	routerGroup.GET("/v1/user/getuserbyemail/:email_user", authAPI.AuthMiddleware("object", "view"), userAPI.GetUserByEmail)

	//SignIN/UP API
	routerGroup.GET("/v1/token", userAPI.UserTokenListHandler)
	routerGroup.GET("/login", userAPI.LoginHandle)
	routerGroup.POST("/signup", userAPI.AddUserSignUpHandeler)
}

func AddUserRouteDev(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	userRepository := repository.UserRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	userAPI := api.UserAPI{
		UserRepository: &userRepository,
	}

	routerGroup.GET("/v1/user", userAPI.UserListHandler)
	routerGroup.POST("/v1/user", userAPI.AddUserHandeler)
	routerGroup.POST("/v1/getroomuser", userAPI.GetUserRoomByUserID)
	routerGroup.POST("/v1/user/updateuser", userAPI.UpdateUserHandler)
	routerGroup.PUT("/v1/user/:user_id", userAPI.EditUserNameHandler)
	routerGroup.DELETE("/v1/user/:user_id", userAPI.DeleteUserByIDHandler)
	routerGroup.GET("/v1/user/getuserbyemail/:email_user", userAPI.GetUserByEmail)

	//SignIN/UP API
	routerGroup.GET("/v1/token", userAPI.UserTokenListHandler)
	routerGroup.GET("/login", userAPI.LoginHandle)
	routerGroup.POST("/signup", userAPI.AddUserSignUpHandeler)
}
