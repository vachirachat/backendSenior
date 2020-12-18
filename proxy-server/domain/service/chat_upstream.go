package service

import (
	"backendSenior/domain/model"
	"encoding/json"
	"fmt"
	"proxySenior/domain/interface/repository"
)

// ChatUpstreamService manages sending message to controller
type ChatUpstreamService struct {
	upstream   repository.UpstreamMessageRepository
	encryption *EncryptionService
}

// NewChatUpstreamService create instance of ChatUpstreamService
func NewChatUpstreamService(controller repository.UpstreamMessageRepository, encryption *EncryptionService) *ChatUpstreamService {
	return &ChatUpstreamService{
		upstream:   controller,
		encryption: encryption,
	}
}

// SendMessage encrypt mesasge and forward to upstream
func (service *ChatUpstreamService) SendMessage(message model.Message) error {
	encryptedMessage, err := service.encryption.Encrypt(message)
	if err != nil {
		fmt.Printf("send error: can't encrypt: %s\n", err.Error())
		return err
	}
	data, err := json.Marshal(encryptedMessage)
	if err != nil {
		return err
	}
	fmt.Printf("[upstream] --> %+v\n", encryptedMessage)
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
