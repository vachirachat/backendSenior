package route

import (
	"fmt"
	"net/http"
	"proxySenior/domain/service"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// MessageRouteHandler handle route for getting past message
type MessageRouteHandler struct {
	messageService *service.MessageService
	// TODO auth
}

// NewMessageRouteHandler create new route handler
func NewMessageRouteHandler(messageService *service.MessageService) *MessageRouteHandler {
	return &MessageRouteHandler{
		messageService: messageService,
	}
}

// Mount make handler handle request on that path
func (handler *MessageRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("/", handler.getMessagesHandler)
}

func (handler *MessageRouteHandler) getMessagesHandler(context *gin.Context) {

	roomID := context.Query("roomId")
	if roomID == "" || !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad query"})
		return
	}

	messages, err := handler.messageService.GetMessageForRoom(roomID, nil)

	if err != nil {
		fmt.Println("err message", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, messages)
}
