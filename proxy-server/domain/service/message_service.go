package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
)

// MessageService i sservice for getting message from controller and decrypt it for user
type MessageService struct {
	messageRepo repository.MessageRepository
	encryption  *EncryptionService
}

// NewMessageService create new instance of message service
func NewMessageService(messageRepo repository.MessageRepository, enc *EncryptionService) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		encryption:  enc,
	}
}

func (service *MessageService) GetMessageForRoom(roomID string, timeRange *model.TimeRange) ([]model.Message, error) {
	messages, err := service.messageRepo.GetMessagesByRoom(roomID, timeRange)
	if err != nil {
		return nil, err
	}

	decrypted := make([]model.Message, len(messages))
	for i := 0; i < len(messages); i++ {
		m, err := service.encryption.Decrypt(messages[i])
		if err != nil {
			decrypted = decrypted[:i]
			return decrypted, err
		}
		decrypted[i] = m
	}
	return decrypted, nil
}
