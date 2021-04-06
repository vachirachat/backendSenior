package route

import (
	"fmt"
	"net/http"
	"proxySenior/domain/encryption"
	model_proxy "proxySenior/domain/model"
	"proxySenior/domain/service"
	"proxySenior/domain/service/key_service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// MessageRouteHandler handle route for getting past message
type MessageRouteHandler struct {
	messageService *service.MessageService
	key            *key_service.KeyService
	// TODO auth
}

// NewMessageRouteHandler create new route handler
func NewMessageRouteHandler(messageService *service.MessageService, key *key_service.KeyService) *MessageRouteHandler {
	return &MessageRouteHandler{
		messageService: messageService,
		key:            key,
	}
}

// Mount make handler handle request on that path
func (h *MessageRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("/", h.getMessagesHandler)
}

func (h *MessageRouteHandler) getMessagesHandler(context *gin.Context) {

	roomID := context.Query("roomId")
	if roomID == "" || !bson.IsObjectIdHex(roomID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad query"})
		return
	}

	messages, err := h.messageService.GetMessageForRoom(roomID, nil)
	if err != nil {
		fmt.Println("error getting message from controller", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	keys, err := h.getKeyFromRoom(roomID)
	if err != nil {
		fmt.Println("[message route] can't get key to decrypt message", err)
		context.JSON(500, gin.H{"status": "error"})
		return
	}

	for i := range messages {
		cipherText, err := encryption.B64Decode([]byte(messages[i].Data))
		if err != nil {
			messages[i].Data = fmt.Sprintf("b64 decode: %v", err)
			continue
		}
		key := keyFor(keys, messages[i].TimeStamp)
		text, err := encryption.AESDecrypt(cipherText, key)
		if err != nil {
			messages[i].Data = fmt.Sprintf("aes decrypt: %v", err)
			continue
		}
		messages[i].Data = string(text)
	}

	context.JSON(http.StatusOK, messages)
}

//getKeyFromRoom determine where to get key and get the key
func (h *MessageRouteHandler) getKeyFromRoom(roomID string) ([]model_proxy.KeyRecord, error) {
	local, err := h.key.IsLocal(roomID)
	if err != nil {
		return nil, fmt.Errorf("error deftermining locality ok key %v", err)
	}

	var keys []model_proxy.KeyRecord
	if local {
		//fmt.Println("[message] use LOCAL key for", roomID)
		keys, err = h.key.GetKeyLocal(roomID)
		if err != nil {
			return nil, fmt.Errorf("error getting key locally %v", err)
		}
	} else {
		//fmt.Println("[message] use REMOTE key for room", roomID)
		keys, err = h.key.GetKeyRemote(roomID)
		if err != nil {
			return nil, fmt.Errorf("error getting key remotely %v", err)
		}
	}

	return keys, nil
}

// keyFor is helper function for finding key in array by time
func keyFor(keys []model_proxy.KeyRecord, timestamp time.Time) []byte {
	var key []byte
	found := false
	for _, k := range keys {
		if k.ValidFrom.Before(timestamp) && (k.ValidTo.IsZero() || k.ValidTo.After(timestamp)) {
			key = k.Key
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	return key
}
