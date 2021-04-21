package service

import (
	"backendSenior/domain/interface/repository"

	"backendSenior/domain/model"
)

// MessageService message service provide access to message related functions
type MessageService struct {
	messageRepo repository.MessageRepository
}

// NewMessageService create message service from repository
func NewMessageService(msgRepo repository.MessageRepository) *MessageService {
	return &MessageService{
		messageRepo: msgRepo,
	}
}

func (service *MessageService) GetAllMessages() ([]model.Message, error) {
	messages, err := service.messageRepo.GetAllMessages(nil)
	return messages, err
}

func (service *MessageService) GetMessageByRoom(roomID string) ([]model.Message, error) {
	messages, err := service.messageRepo.GetMessagesByRoom(roomID, nil)
	return messages, err
}

func (service *MessageService) GetMessageByRoomWithTimeRange(roomID string, timeRange *model.TimeRange) ([]model.Message, error) {
	messages, err := service.messageRepo.GetMessagesByRoom(roomID, timeRange)
	return messages, err
}

func (service *MessageService) GetMessageByID(messageId string) (model.Message, error) {
	msg, err := service.messageRepo.GetMessageByID(messageId)
	return msg, err
}

func (service *MessageService) AddMessage(newMessage model.Message) (string, error) {
	msgID, err := service.messageRepo.AddMessage(newMessage)
	return msgID, err
}

func (service *MessageService) DeleteMessageByID(messageId string) error {
	err := service.messageRepo.DeleteMessageByID(messageId)
	return err
}
