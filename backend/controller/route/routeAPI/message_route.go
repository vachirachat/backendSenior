package routeAPI

import (
	mongo_repository "backendSenior/data/repository/mongo_repository"
	"backendSenior/domain/interface/repository"
	service "backendSenior/domain/usecase"
	"backendSenior/domain/usecase/auth"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

// Todo move this else where
func AddMessageRoute(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	var messageRepository repository.MessageRepository = mongo_repository.MessageRepositoryMongo{
		ConnectionDB: connectionDB,
	}

	var userRepository repository.UserRepository = mongo_repository.UserRepositoryMongo{
		ConnectionDB: connectionDB,
	}

	messageService := service.NewMessageService(messageRepository)
	authService := &auth.AuthService{
		UserRepository: userRepository,
	}

	msg := NewMessageRouteHandler(messageService, authService)

	subGroup := routerGroup.Group("/v1/message")
	msg.Mount(subGroup)

}

// func AddMessageRouteDev(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

// 	//Message API
// 	messageRepository := repository.MessageRepositoryMongo{
// 		ConnectionDB: connectionDB,
// 	}
// 	messageAPI := service.MessageService{
// 		MessageRepository: &messageRepository,
// 	}

// 	// Autherize Middleware API
// 	routerGroup.GET("/v1/message", messageAPI.MessageListHandler)
// 	routerGroup.POST("/v1/message", messageAPI.AddMessageHandeler)
// 	// route.PUT("/v1/message/:message_id", authAPI.AuthMiddleware("object", "view") ,messageAPI.EditMessageHandler)
// 	routerGroup.DELETE("/v1/message/:message_id", messageAPI.DeleteMessageByIDHandler)

// }
