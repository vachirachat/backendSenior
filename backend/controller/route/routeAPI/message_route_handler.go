package routeAPI

import (
	"backendSenior/domain/model"
	service "backendSenior/domain/usecase"
	"backendSenior/domain/usecase/auth"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MessageRouteHandler is Handler (controller) for message related route
type MessageRouteHandler struct {
	messageService *service.MessageService
	authService    *auth.AuthService
}

// NewMessageRouteHandler create handler for message route
func NewMessageRouteHandler(msgService *service.MessageService, authService *auth.AuthService) *MessageRouteHandler {
	return &MessageRouteHandler{
		messageService: msgService,
		authService:    authService,
	}
}

//Mount make messageRouteHandler handler request from specific `RouterGroup`
func (handler *MessageRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("/", handler.authService.AuthMiddleware("object", "view"), handler.MessageListHandler)
	routerGroup.POST("/", handler.authService.AuthMiddleware("object", "view"), handler.AddMessageHandeler)
	// route.PUT("/message/:message_id", handler.authService.AuthMiddleware("object", "view") ,handler.EditMessageHandler)
	routerGroup.DELETE("/:message_id", handler.authService.AuthMiddleware("object", "view"), handler.DeleteMessageByIDHandler)
}

// MessageListHandler return all messages
func (handler *MessageRouteHandler) MessageListHandler(context *gin.Context) {
	// return value
	var messagesInfo model.MessageInfo

	messages, err := handler.messageService.GetAllMessages()
	if err != nil {
		log.Println("error MessageListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	messagesInfo.Messages = messages
	context.JSON(http.StatusOK, messagesInfo)
}

// GetMessageByIDHandler return message by Id
func (handler *MessageRouteHandler) GetMessageByIDHandler(context *gin.Context) {
	messageID := context.Param("message_id")

	message, err := handler.messageService.GetMessageByID(messageID)
	if err != nil {
		log.Println("error GetMessageByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, message)
}

// AddMessageHandeler
func (handler *MessageRouteHandler) AddMessageHandeler(context *gin.Context) {
	var message model.Message

	err := context.ShouldBindJSON(&message)
	if err != nil {
		log.Println("error AddMessageHandeler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}
	err = handler.messageService.AddMessage(message)
	if err != nil {
		log.Println("error AddMessageHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (handler *MessageRouteHandler) DeleteMessageByIDHandler(context *gin.Context) {
	messageID := context.Param("message_id")
	err := handler.messageService.DeleteMessageByID(messageID)
	if err != nil {
		log.Println("error DeleteMessageHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(http.StatusNoContent, gin.H{"status": "success"})
}
