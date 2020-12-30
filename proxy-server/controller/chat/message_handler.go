package chat

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	chatmodel "backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/message_types"
	"backendSenior/domain/model/chatsocket/room"
	"encoding/json"
	"fmt"
	"proxySenior/domain/service"
)

// MessageHandler handle message from controller
// ex. broadcasting to users, updating room-user repo
type MessageHandler struct {
	upstreamService   *service.ChatUpstreamService   // recv message from controlller
	downstreamService *service.ChatDownstreamService // bcast message to user
	roomUserRepo      repository.RoomUserRepository  // update room on event from controller
}

// NewMessageHandler creates new MessageHandler
func NewMessageHandler(upstream *service.ChatUpstreamService, downstream *service.ChatDownstreamService, roomUserRepo repository.RoomUserRepository) *MessageHandler {
	return &MessageHandler{
		upstreamService:   upstream,
		downstreamService: downstream,
		roomUserRepo:      roomUserRepo,
	}
}

func (h *MessageHandler) Start() {
	pipe := make(chan []byte, 100)
	h.upstreamService.RegsiterHandler(pipe)
	defer h.upstreamService.UnRegsiterHandler(pipe)

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
			fmt.Println("The message is", msg)

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
