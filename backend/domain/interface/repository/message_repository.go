package repository

import (
	"backendSenior/domain/model"
)

// MessageRepository defines interface for Message Repositories
type MessageRepository interface {
	GetAllMessages() ([]model.Message, error)
	// GetLastMessage() (model.Message, error)
	GetMessageByID(userID string) (model.Message, error)
	AddMessage(message model.Message) error
	DeleteMessageByID(userID string) error
}
