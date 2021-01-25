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

// OnConnect register channel to be notified when upstream is connected
func (service *ChatUpstreamService) OnConnect(channel chan struct{}) {
	service.upstream.OnConnect(channel)
	// service.up = append(upstream.onConnectRecv, channel)
}

// OffConnect unregister channel from being notified when upstream is connected
func (service *ChatUpstreamService) OffConnect(channel chan struct{}) {
	service.upstream.OffConnect(channel)
}

// OnDisconnect register channel to be notified when upstream is disconnected
func (service *ChatUpstreamService) OnDisconnect(channel chan struct{}) {
	service.upstream.OnDisconnect(channel)
}

// OffDisconnect unregister channel from being notified when upstream is disconnected
func (service *ChatUpstreamService) OffDisconnect(channel chan struct{}) {
	service.upstream.OffDisconnect(channel)
}
