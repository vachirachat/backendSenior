package route

import (
	"backendSenior/domain/model"
	"backendSenior/domain/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MessageRouteHandler is Handler (controller) for message related route
type MessageRouteHandler struct {
	messageService *service.MessageService
}

// NewMessageRouteHandler create handler for message route
func NewMessageRouteHandler(msgService *service.MessageService) *MessageRouteHandler {
	return &MessageRouteHandler{
		messageService: msgService,
	}
}

//Mount make messageRouteHandler handler request from specific `RouterGroup`
func (handler *MessageRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("/" /*handler.authService.AuthMiddleware("object", "view")*/, handler.messageListHandler)
	routerGroup.POST("/" /*handler.authService.AuthMiddleware("object", "view")*/, handler.addMessageHandeler)
	// route.PUT("/message/:message_id" /*handler.authService.AuthMiddleware("object", "view")*/ ,handler.editMessageHandler)
	routerGroup.DELETE("/:message_id" /*handler.authService.AuthMiddleware("object", "view")*/, handler.deleteMessageByIDHandler)
}

// MessageListHandler return all messages
func (handler *MessageRouteHandler) messageListHandler(context *gin.Context) {
	// return value
	var messagesInfo model.MessagesResponse

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
func (handler *MessageRouteHandler) getMessageByIDHandler(context *gin.Context) {
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
func (handler *MessageRouteHandler) addMessageHandeler(context *gin.Context) {
	var message model.Message

	err := context.ShouldBindJSON(&message)
	if err != nil {
		log.Println("error AddMessageHandeler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}
	_, err = handler.messageService.AddMessage(message)

	if err != nil {
		log.Println("error AddMessageHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (handler *MessageRouteHandler) deleteMessageByIDHandler(context *gin.Context) {
	messageID := context.Param("message_id")
	err := handler.messageService.DeleteMessageByID(messageID)
	if err != nil {
		log.Println("error DeleteMessageHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(http.StatusNoContent, gin.H{"status": "success"})
}
