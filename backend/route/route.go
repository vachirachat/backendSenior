package route

import (
	"backendSenior/api"
	"backendSenior/repository"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

func NewRouteProduct(route *gin.Engine, connectionDB *mgo.Session) {

	userRepository := repository.UserRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	userAPI := api.UserAPI{
		UserRepository: &userRepository,
	}
	route.GET("api/v1/user", userAPI.UserListHandler)
	route.POST("api/v1/user", userAPI.AddUserHandeler)
	route.PUT("api/v1/user/:user_id", userAPI.EditUserNameHandler)
	route.DELETE("api/v1/user/:user_id", userAPI.DeleteUserByIDHandler)

	//Token
	route.GET("api/v1/token", userAPI.UserTokenListHandler)
	route.GET("login", userAPI.LoginHandle)
	route.POST("signUp", userAPI.AddUserSignUpHandeler)
	//	route.GET("signUp", userAPI.signUp)
}
