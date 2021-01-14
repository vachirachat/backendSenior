package service

import (
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/message_types"
	"encoding/json"
	"fmt"
	"proxySenior/domain/interface/repository"
)

// ChatUpstreamService manages sending message to controller
type ChatUpstreamService struct {
	upstream repository.UpstreamMessageRepository
}

// NewChatUpstreamService create instance of ChatUpstreamService
func NewChatUpstreamService(controller repository.UpstreamMessageRepository) *ChatUpstreamService {
	return &ChatUpstreamService{
		upstream: controller,
	}
}

// SendMessage send message to controller, it doesn't encrypt message
func (service *ChatUpstreamService) SendMessage(message model.Message) error {
	data, err := json.Marshal(chatsocket.Message{
		Type:    message_types.Chat,
		Payload: message,
	})
	if err != nil {
		return err
	}
	fmt.Printf("[upstream] --> %+v\n", data)
	err = service.upstream.SendMessage(data)
	return err
}

// RegsiterHandler add channel to be notified when message is received
func (service *ChatUpstreamService) RegsiterHandler(channel chan []byte) error {
	return service.upstream.RegisterHandler(channel)
}

// UnRegsiterHandler remove channel from being notified when message is received
func (service *ChatUpstreamService) UnRegsiterHandler(channel chan []byte) error {
	return service.upstream.UnRegisterHandler(channel)
}
