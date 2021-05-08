package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/dto"
	"backendSenior/domain/model"
	"backendSenior/domain/service"
	"backendSenior/utills"
	g "common/utils/ginutils"
	"errors"
	"log"
	"net/http"

	"time"

	"github.com/ahmetb/go-linq/v3"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// MessageRouteHandler is Handler (controller) for message related route
type MessageRouteHandler struct {
	messageService *service.MessageService
	fileService    *service.FileService
	roomService    *service.RoomService
	auth           *auth.JWTMiddleware
	proxyMw        *auth.ProxyMiddleware
	validate       *utills.StructValidator
}

// NewMessageRouteHandler create handler for message route
func NewMessageRouteHandler(
	msgService *service.MessageService,
	fileService *service.FileService,
	roomService *service.RoomService,
	auth *auth.JWTMiddleware,
	proxyMw *auth.ProxyMiddleware,
	validate *utills.StructValidator,
) *MessageRouteHandler {
	return &MessageRouteHandler{
		messageService: msgService,
		fileService:    fileService,
		roomService:    roomService,
		auth:           auth,
		proxyMw:        proxyMw,
		validate:       validate,
	}
}

//Mount make messageRouteHandler handler request from specific `RouterGroup`
func (handler *MessageRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("/", handler.proxyMw.AlternativeAuth(), handler.auth.AuthRequired("user", "view"), handler.auth.IsInRoomMiddleWareQuery("roomId"), g.InjectGin(handler.messageListHandler))
	routerGroup.POST("/find", handler.proxyMw.AlternativeAuth(), handler.auth.AuthRequired("user", "edit"), g.InjectGin(handler.findMessage)) // GET upto last 1000 message

	routerGroup.POST("/" /*handler.authService.AuthMiddleware("object", "view")*/, handler.addMessageHandeler)
	// route.PUT("/message/:message_id" /*handler.authService.AuthMiddleware("object", "view")*/ ,handler.editMessageHandler)
	routerGroup.DELETE("/:message_id", handler.auth.AuthRequired("user", "edit"), g.InjectGin(handler.deleteMessageByIDHandler))
}

// MessageListHandler return all messages or message for specified room
func (handler *MessageRouteHandler) messageListHandler(context *gin.Context, req struct{}) error {
	// return value
	var messagesInfo model.MessagesResponse

	roomID := context.Query("roomId")
	if roomID != "" && !bson.IsObjectIdHex(roomID) {
		return g.NewError(400, "bad roomID param")
	}

	var messages []model.Message
	var err error

	if roomID != "" {
		messages, err = handler.messageService.GetMessageByRoom(roomID)
	} else {
		messages, err = handler.messageService.GetAllMessages()
	}

	if err != nil {
		return err
	}
	messagesInfo.Messages = messages
	context.JSON(http.StatusOK, messagesInfo)
	return nil
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
		return
	}
	_, err = handler.messageService.AddMessage(message)
	if err != nil {
		log.Println("error AddMessageHandeler", err.Error())
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
		return err
	}
	context.JSON(200, g.Response{Success: true, Message: "deleted file"})
	return nil
}

func (handler *MessageRouteHandler) findMessage(c *gin.Context, req struct {
	Body dto.FindMessageDto
}) error {
	body := req.Body

	if body.To.IsZero() {
		body.To = time.Now()
	}
	if body.From.IsZero() {
		body.From = time.Unix(0, 0)
	}
	if body.From.After(body.To) {
		return g.NewError(400, "body.From must be before body.To")
	}

	msgs, err := handler.messageService.GetMessageByRoomWithTimeRange(body.RoomID.Hex(), &model.TimeRange{
		From: time.Time{},
		To:   time.Time{},
	})
	if err != nil {
		return err
	}
	c.JSON(200, msgs)
	return nil
}
