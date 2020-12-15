package service

import (
	"backendSenior/domain/model"
	"encoding/json"
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
	encryptedMessage := service.encryption.Encrypt(message)
	data, err := json.Marshal(encryptedMessage)
	if err != nil {
		return err
	}
	err = service.upstream.SendMessage(data)
	return err
}

func (service *ChatUpstreamService) RegsiterHandler(channel chan []byte) error {
	return service.upstream.RegisterHandler(channel)
}

func (service *ChatUpstreamService) UnRegsiterHandler(channel chan []byte) error {
	return service.upstream.UnRegisterHandler(channel)
}
