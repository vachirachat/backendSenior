package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/model"
	"backendSenior/domain/service"
	"backendSenior/utills"
	g "common/utils/ginutils"
	"errors"
	"log"
	"net/http"

	"github.com/ahmetb/go-linq/v3"
	"github.com/globalsign/mgo"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// MessageRouteHandler is Handler (controller) for message related route
type MessageRouteHandler struct {
	messageService *service.MessageService
	fileService    *service.FileService
	roomService    *service.RoomService
	auth           *auth.JWTMiddleware
	validate       *utills.StructValidator
}

// NewMessageRouteHandler create handler for message route
func NewMessageRouteHandler(
	msgService *service.MessageService,
	fileService *service.FileService,
	roomService *service.RoomService,
	auth *auth.JWTMiddleware,
	validate *utills.StructValidator,
) *MessageRouteHandler {
	return &MessageRouteHandler{
		messageService: msgService,
		fileService:    fileService,
		roomService:    roomService,
		auth:           auth,
		validate:       validate,
	}
}

//Mount make messageRouteHandler handler request from specific `RouterGroup`
func (handler *MessageRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	// routerGroup.GET("/" /*handler.authService.AuthMiddleware("object", "view")*/, handler.messageListHandler)
	routerGroup.POST("/" /*handler.authService.AuthMiddleware("object", "view")*/, handler.addMessageHandeler)
	// route.PUT("/message/:message_id" /*handler.authService.AuthMiddleware("object", "view")*/ ,handler.editMessageHandler)
	routerGroup.DELETE("/:message_id", handler.auth.AuthRequired(), g.InjectGin(handler.deleteMessageByIDHandler))
}

// MessageListHandler return all messages
func (handler *MessageRouteHandler) messageListHandler(context *gin.Context) {
	// return value
	var messagesInfo model.MessagesResponse

	roomID := context.Query("roomId")
	if roomID != "" && !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad query"})
		return
	}

	var messages []model.Message
	var err error

	if roomID != "" {
		messages, err = handler.messageService.GetMessageByRoom(roomID)
	} else {
		messages, err = handler.messageService.GetAllMessages()
	}

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

func (handler *MessageRouteHandler) deleteMessageByIDHandler(context *gin.Context, req struct{}) error {
	messageID := context.Param("message_id")
	if !bson.IsObjectIdHex(messageID) {
		return g.NewError(400, "invalid room ID")
	}

	msg, err := handler.messageService.GetMessageByID(messageID)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, "message not found")
		}
	}
	userID := bson.ObjectIdHex(context.GetString(auth.UserIdField))
	if msg.UserID != userID {
		return g.NewError(403, "not your message")
	}
	room, err := handler.roomService.GetRoomByID(msg.RoomID.Hex())
	if err != nil {
		return err
	}
	if !linq.From(room.ListUser).Contains(userID) {
		return g.NewError(403, "not in the room")
	}

	if msg.Type == model.MsgFile {
		if err := handler.fileService.DeleteFile(msg.FileID); err != nil && !errors.Is(err, mgo.ErrNotFound) {
			return err
		}
	} else if msg.Type == model.MsgImage {
		if err := handler.fileService.DeleteImage(msg.FileID); err != nil && !errors.Is(err, mgo.ErrNotFound) {
			return err
		}
	}

	if err := handler.messageService.DeleteMessageByID(messageID); err != nil {
		log.Println("error DeleteMessageHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(200, g.Response{Success: true, Message: "deleted file"})
	return nil
}
