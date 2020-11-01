package route

import (
	"backendSenior/api"
	"backendSenior/api/auth"
	"backendSenior/repository"
	"net/http"

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
	route.GET("api/v1/user/:email_user", userAPI.GetUserByEmail)

	// Autherize Middleware API
	authAPI := auth.Auth{
		UserRepository: &userRepository,
	}

	//Google oAuth
	route.LoadHTMLGlob("route/*")
	route.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	// OauthGoogle
	route.GET("/auth/google/login", auth.OauthGoogleLogin)
	route.GET("/auth/google/callback", auth.OauthGoogleCallback)

	//SignIN/UP API
	route.GET("api/v1/token", userAPI.UserTokenListHandler)
	route.GET("login", userAPI.LoginHandle)
	route.POST("signup", userAPI.AddUserSignUpHandeler)

	// Room API
	roomRepository := repository.RoomRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	roomAPI := api.RoomAPI{
		RoomRepository: &roomRepository,
	}
	route.GET("api/v1/room", authAPI.AuthMiddleware("resource", "scope"), roomAPI.RoomListHandler)
	route.POST("api/v1/room", roomAPI.AddRoomHandeler)
	route.PUT("api/v1/room/:room_id", roomAPI.EditRoomNameHandler)
	route.DELETE("api/v1/room/:room_id", roomAPI.DeleteRoomByIDHandler)

	//Message API
	messageRepository := repository.MessageRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	messageAPI := api.MessageAPI{
		MessageRepository: &messageRepository,
	}
	route.GET("api/v1/message", messageAPI.MessageListHandler)
	route.POST("api/v1/message", messageAPI.AddMessageHandeler)
	// route.PUT("api/v1/message/:message_id", messageAPI.EditMessageHandler)
	route.DELETE("api/v1/message/:message_id", messageAPI.DeleteMessageByIDHandler)

}
