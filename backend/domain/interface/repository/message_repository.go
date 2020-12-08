package repository

import (
	"backendSenior/domain/model"
)

// MessageRepository defines interface for Message Repositories
type MessageRepository interface {
	GetAllMessages(timeRange *model.TimeRange) ([]model.Message, error)
	GetMessageByRoom(roomID string, timeRange *model.TimeRange)
	// GetLastMessage() (model.Message, error)
	GetMessageByID(userID string) (model.Message, error)
	AddMessage(message model.Message) error
	DeleteMessageByID(userID string) error
}
