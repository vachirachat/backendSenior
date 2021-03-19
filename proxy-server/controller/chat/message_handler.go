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
	"proxySenior/domain/plugin"
	"proxySenior/domain/service"
)

// MessageHandler handle message from controller
// ex. broadcasting to users, updating room-user repo
type MessageHandler struct {
	upstreamService     *service.ChatUpstreamService   // recv message from controlller
	downstreamService   *service.ChatDownstreamService // bcast message to user
	roomUserRepo        repository.RoomUserRepository  // update room on event from controller
	encryption          *service.EncryptionService     // for decrypting message
	onMessagePortPlugin *plugin.OnMessagePortPlugin
}

// NewMessageHandler creates new MessageHandler
func NewMessageHandler(upstream *service.ChatUpstreamService, downstream *service.ChatDownstreamService, roomUserRepo repository.RoomUserRepository, encryption *service.EncryptionService, onMessagePortPlugin *plugin.OnMessagePortPlugin) *MessageHandler {
	return &MessageHandler{
		upstreamService:     upstream,
		downstreamService:   downstream,
		roomUserRepo:        roomUserRepo,
		encryption:          encryption,
		onMessagePortPlugin: onMessagePortPlugin,
	}
}

func (h *MessageHandler) Start() {
	pipe := make(chan []byte, 100)
	h.upstreamService.RegsiterHandler(pipe)
	defer h.upstreamService.UnRegsiterHandler(pipe)
	log.Println("Start", "Chat Service")
	for {
		data := <-pipe
		fmt.Printf("[upstream] <-- %s\n", data)

		var rawMessage chatmodel.RawMessage
		err := json.Unmarshal(data, &rawMessage)

		if err != nil {
			fmt.Println("error parsing message from upstream", err)
			fmt.Printf("the message was [%s]\n", data)
			continue
		}

		if rawMessage.Type == message_types.Chat {
			var msg model.Message
			err := json.Unmarshal(rawMessage.Payload, &msg)
			if err != nil {
				fmt.Println("error parsing message *payload* from upstream", err)
				fmt.Printf("the message was [%s]\n", data)
				continue
			}
			//  Task: Plugin-Encryption : Forward to Decryption
			msg, err = h.encryption.DecryptController(msg)
			if err != nil {
				fmt.Println("Error decrpyting", err)
				continue
			}
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
				fmt.Printf("the message was [%s]\n", data)
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

		}

	}

}
