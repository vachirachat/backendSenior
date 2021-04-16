package chat

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	chatmodel "backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/message_types"
	"backendSenior/domain/model/chatsocket/room"
	"encoding/json"
	"fmt"
	"proxySenior/domain/plugin"
	"proxySenior/domain/service"
	"proxySenior/domain/service/key_service"
	"time"
)

// MessageHandler handle message from controller
// ex. broadcasting to users, updating room-user repo
type MessageHandler struct {
	upstreamService     *service.ChatUpstreamService   // recv message from controlller
	downstreamService   *service.ChatDownstreamService // bcast message to user
	roomUserRepo        repository.RoomUserRepository  // update room on event from controller
	key                 *key_service.KeyService        // for getting key to decrypt the messages
	onMessagePortPlugin *plugin.OnMessagePortPlugin
	encryption          *service.EncryptionService
}

// NewMessageHandler creates new MessageHandler
func NewMessageHandler(upstream *service.ChatUpstreamService, downstream *service.ChatDownstreamService, roomUserRepo repository.RoomUserRepository, key *key_service.KeyService, onMessagePortPlugin *plugin.OnMessagePortPlugin, enc *service.EncryptionService) *MessageHandler {
	return &MessageHandler{
		upstreamService:     upstream,
		downstreamService:   downstream,
		roomUserRepo:        roomUserRepo,
		onMessagePortPlugin: onMessagePortPlugin,
		key:                 key,
		encryption:          enc,
	}
}

// Start listen message from upstream
func (h *MessageHandler) Start() {
	pipe := make(chan []byte, 200)
	h.upstreamService.RegisterHandler(pipe)
	defer h.upstreamService.UnRegisterHandler(pipe)

	for {
		incMessage := <-pipe
		var rawMessage chatmodel.RawMessage
		err := json.Unmarshal(incMessage, &rawMessage)
		if err != nil {
			fmt.Println("error parsing message from upstream", err)
			fmt.Printf("the message was [%s]\n", incMessage)
			continue
		}

		if rawMessage.Type == message_types.Chat {
			var msg model.Message
			if err := json.Unmarshal(rawMessage.Payload, &msg); err != nil {
				fmt.Println("error parsing message *payload* from upstream", err)
				fmt.Printf("the message was [%s]\n", incMessage)
				continue
			}
			// decrypt message (either by plugin or
			if err := h.encryption.DecryptController(&msg); err != nil {
				fmt.Printf("room %s, msgId %s, decrypt error: %s\n", msg.RoomID.Hex(), msg.MessageID.Hex(), err)
				continue
			}

			if h.onMessagePortPlugin != nil && h.onMessagePortPlugin.IsEnabled() {
				err := h.onMessagePortPlugin.OnMessagePortPlugin(msg)
				fmt.Println("[plugin] called on message error", err)
			}
			err = h.downstreamService.BroadcastMessageToRoom(msg.RoomID.Hex(), msg)
			if err != nil {
				fmt.Println("Error BCasting", err)
			}

		} else if rawMessage.Type == message_types.Room {
			var event room.MemberEvent
			err = json.Unmarshal(rawMessage.Payload, &event)
			if err != nil {
				fmt.Println("error parsing room event *payload* from upstream", err)
				fmt.Printf("the message was [%s]\n", incMessage)
				continue
			}

			if event.Type == room.Join {
				fmt.Printf("[handle room event] %s JOIN %s\n", event.RoomID, event.Members)
				h.roomUserRepo.AddUsersToRoom(event.RoomID, event.Members)
			} else if event.Type == room.Leave {
				fmt.Printf("[handle room event] %s LEAVE %s\n", event.RoomID, event.Members)
				h.roomUserRepo.RemoveUsersFromRoom(event.RoomID, event.Members)
			} else {
				fmt.Println("[handle room event] unkown event type", event.Type)
			}
		} else if rawMessage.Type == message_types.InvalidateMaster {
			// TODO: this is work around to wait master to disconnect, ensuring that master is changed
			time.Sleep(100 * time.Millisecond)
			var roomID string
			err := json.Unmarshal(rawMessage.Payload, &roomID)
			if err != nil {
				fmt.Println("bad payload", err)
				continue
			}
			fmt.Println("invalidate master", roomID)
			h.key.RevalidateRoomMaster(roomID)
		} else if rawMessage.Type == message_types.InvalidateKey {
			var roomID string
			err := json.Unmarshal(rawMessage.Payload, &roomID)
			if err != nil {
				fmt.Println("bad payload", err)
				continue
			}
			fmt.Println("invalidate KEY", roomID)
			h.key.RevalidateKeyCache(roomID)
		}

	}
}
