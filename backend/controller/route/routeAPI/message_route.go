package routeAPI

import (
	"backendSenior/data/repository"
	api "backendSenior/domain/usecase"
	"backendSenior/domain/usecase/auth"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

func AddMessageRoute(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	//Message API
	messageRepository := repository.MessageRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	messageAPI := api.MessageAPI{
		MessageRepository: &messageRepository,
	}

	userRepository := repository.UserRepositoryMongo{
		ConnectionDB: connectionDB,
	}

	// Autherize Middleware API
	authAPI := auth.Auth{
		UserRepository: &userRepository,
	}
	routerGroup.GET("/v1/message", authAPI.AuthMiddleware("object", "view"), messageAPI.MessageListHandler)
	routerGroup.POST("/v1/message", authAPI.AuthMiddleware("object", "view"), messageAPI.AddMessageHandeler)
	// route.PUT("/v1/message/:message_id", authAPI.AuthMiddleware("object", "view") ,messageAPI.EditMessageHandler)
	routerGroup.DELETE("/v1/message/:message_id", authAPI.AuthMiddleware("object", "view"), messageAPI.DeleteMessageByIDHandler)

}

func AddMessageRouteDev(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	//Message API
	messageRepository := repository.MessageRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	messageAPI := api.MessageAPI{
		MessageRepository: &messageRepository,
	}

	// Autherize Middleware API
	routerGroup.GET("/v1/message", messageAPI.MessageListHandler)
	routerGroup.POST("/v1/message", messageAPI.AddMessageHandeler)
	// route.PUT("/v1/message/:message_id", authAPI.AuthMiddleware("object", "view") ,messageAPI.EditMessageHandler)
	routerGroup.DELETE("/v1/message/:message_id", messageAPI.DeleteMessageByIDHandler)

}
