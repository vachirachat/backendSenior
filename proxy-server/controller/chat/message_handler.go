package chat

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	chatmodel "backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/message_types"
	"backendSenior/domain/model/chatsocket/room"
	"encoding/json"
	"fmt"
	"log"
	"proxySenior/domain/encryption"
	model_proxy "proxySenior/domain/model"
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
}

// NewMessageHandler creates new MessageHandler
func NewMessageHandler(upstream *service.ChatUpstreamService, downstream *service.ChatDownstreamService, roomUserRepo repository.RoomUserRepository, key *key_service.KeyService, onMessagePortPlugin *plugin.OnMessagePortPlugin) *MessageHandler {
	return &MessageHandler{
		upstreamService:     upstream,
		downstreamService:   downstream,
		roomUserRepo:        roomUserRepo,
		onMessagePortPlugin: onMessagePortPlugin,
		key:                 key,
	}
}

// Start listen message from upstream
func (h *MessageHandler) Start() {
	pipe := make(chan []byte, 100)
	h.upstreamService.RegsiterHandler(pipe)
	defer h.upstreamService.UnRegsiterHandler(pipe)

	for {
		incMessage := <-pipe
		fmt.Printf("[upstream] <-- %s\n", incMessage)
		var rawMessage chatmodel.RawMessage
		err := json.Unmarshal(incMessage, &rawMessage)
		if err != nil {
			fmt.Println("error parsing message from upstream", err)
			fmt.Printf("the message was [%s]\n", incMessage)
			continue
		}

		if rawMessage.Type == message_types.Chat {
			var msg model.Message
			err := json.Unmarshal(rawMessage.Payload, &msg)
			if err != nil {
				fmt.Println("error parsing message *payload* from upstream", err)
				fmt.Printf("the message was [%s]\n", incMessage)
				continue
			}

			keys, err := h.getKeyFromRoom(msg.RoomID.Hex())
			key := keyFor(keys, msg.TimeStamp)

			encrypted, err := encryption.B64Decode([]byte(msg.Data))
			if err != nil {
				log.Println("error b64 decode message:", err)
				continue
			}
			msgData, err := encryption.AESDecrypt(encrypted, key)
			if err != nil {
				log.Println("error decrypting message:", err)
				continue
			}
			msg.Data = string(msgData)
			fmt.Println("The decrypted message is", msg)

			fmt.Println("try call on message", h.onMessagePortPlugin, h.onMessagePortPlugin.IsEnabled())
			if h.onMessagePortPlugin != nil && h.onMessagePortPlugin.IsEnabled() {
				err := h.onMessagePortPlugin.OnMessagePortPlugin(msg)
				fmt.Println("[plugin] called on message", err)
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

//getKeyFromRoom determine where to get key and get the key
func (h *MessageHandler) getKeyFromRoom(roomID string) ([]model_proxy.KeyRecord, error) {
	local, err := h.key.IsLocal(roomID)
	if err != nil {
		return nil, fmt.Errorf("error deftermining locality ok key %v", err)
	}

	var keys []model_proxy.KeyRecord
	if local {
		fmt.Println("[message] use LOCAL key for", roomID)
		keys, err = h.key.GetKeyLocal(roomID)
		if err != nil {
			return nil, fmt.Errorf("error getting key locally %v", err)
		}
	} else {
		fmt.Println("[message] use REMOTE key for room", roomID)
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
