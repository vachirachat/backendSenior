package repository

import "backendSenior/domain/model"

// SendMessageRepository is used to actually send message
type SendMessageRepository interface {
	SendMessage(connectionID string, data interface{}) error
}

// SocketConnectionRepository used by "Message Mux" to determine where to forward message
type SocketConnectionRepository interface {
	GetConnectionByUser(userID string) ([]string, error)
	AddConnection(conn model.SocketConnection) (string, error) // return generated id of connection
	RemoveConnection(connID string) error
}
