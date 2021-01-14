package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
)

// MessageService is service for getting message from controller and decrypt it for user
type MessageService struct {
	messageRepo repository.MessageRepository
}

// NewMessageService create new instance of message service
func NewMessageService(messageRepo repository.MessageRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
	}
}

// GetMessageForRoom return message from room
func (service *MessageService) GetMessageForRoom(roomID string, timeRange *model.TimeRange) ([]model.Message, error) {
	messages, err := service.messageRepo.GetMessagesByRoom(roomID, timeRange)
	return messages, err
}
