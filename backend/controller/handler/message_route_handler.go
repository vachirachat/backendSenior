package route

import (
	"backendSenior/domain/model"
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"
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
	routerGroup.GET("/" /*handler.authService.AuthMiddleware("object", "view")*/, handler.messageListHandler)
	routerGroup.POST("/" /*handler.authService.AuthMiddleware("object", "view")*/, handler.addMessageHandeler)
	// route.PUT("/message/:message_id" /*handler.authService.AuthMiddleware("object", "view")*/ ,handler.editMessageHandler)
	routerGroup.DELETE("/" /*handler.authService.AuthMiddleware("object", "view")*/, handler.deleteMessageByIDHandler)
	routerGroup.POST("/roommessages" /*handler.authService.AuthMiddleware("object", "view")*/, handler.getMessagesByRoomHandler)
	routerGroup.GET("/getmessagebyid" /*handler.authService.AuthMiddleware("object", "view")*/, handler.getMessageByIDHandler)
}

type roomMessage struct {
	RoomId    string          `json:"roomid" bson:"roomid"`
	TimeRange model.TimeRange `json:"timerange" bson:"timerange"`
}

// getMessageInRoomHandler return all messages
func (handler *MessageRouteHandler) getMessagesByRoomHandler(context *gin.Context) {
	// return value

	var messagesInfo model.MessagesResponse
	var roomMessages roomMessage
	err := context.ShouldBindJSON(&roomMessages)
	if err != nil {
		log.Println("error GetMessageInRoomHandler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}
	// comment: pass roomMessages.TimeRange with address (roomMessages.TimeRange)
	messages, err := handler.messageService.GetMessagesByRoom(roomMessages.RoomId, &roomMessages.TimeRange)

	if err != nil {
		log.Println("error GetMessageInRoomHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	messagesInfo.Messages = messages
	context.JSON(http.StatusOK, messagesInfo)
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
	var messages model.Message
	err := context.ShouldBindJSON(&messages)
	if err != nil {
		log.Println("error GetMessageByIDHandler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	message, err := handler.messageService.GetMessageByID(messages.MessageID.Hex())

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
	var messages model.Message
	err := context.ShouldBindJSON(&messages)
	if err != nil {
		log.Println("error GetMessageByIDHandler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	err = handler.messageService.DeleteMessageByID(messages.MessageID.Hex())
	if err != nil {
		log.Println("error DeleteMessageHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(http.StatusNoContent, gin.H{"status": "success"})
}
