package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"fmt"
	"sync"
)

// ChatDownstreamService manages sending message and connection pool
type ChatDownstreamService struct {
	mapRoom    repository.RoomUserRepository
	send       repository.SendMessageRepository
	mapConn    repository.SocketConnectionRepository
	msgRepo    repository.MessageRepository
	encryption *EncryptionService
}

// NewChatDownstreamService create new instance of chat service
// TODO: this is cheating, since it used same repository as backend
func NewChatDownstreamService(roomUserRepo repository.RoomUserRepository, sender repository.SendMessageRepository, userConnRepo repository.SocketConnectionRepository, msgRepo repository.MessageRepository, encryption *EncryptionService) *ChatDownstreamService {
	return &ChatDownstreamService{
		mapRoom:    roomUserRepo,
		send:       sender,
		mapConn:    userConnRepo,
		msgRepo:    msgRepo,
		encryption: encryption,
	}
}

// SaveMessage save speicified message to repository, returning the ID of message
func (chat *ChatDownstreamService) SaveMessage(message model.Message) (string, error) {
	id, err := chat.msgRepo.AddMessage(message)
	return id, err
}

// SendMessageToConnection send message to specific connection, data will be marshalled
func (chat *ChatDownstreamService) SendMessageToConnection(connID string, message interface{}) error {
	return chat.send.SendMessage(connID, message)
}

// BroadcastMessageToRoom send message to socket of all users in the room
// []byte will be sent as is, but other value will be marshalled
// TODO: in the future there should be broadcast event etc.
func (chat *ChatDownstreamService) BroadcastMessageToRoom(roomID string, message model.Message) error {
	message, err := chat.encryption.Decrypt(message)
	if err != nil {
		fmt.Printf("recv error: can't decrypt: %s\n", err.Error())
		return err
	}

	userIDs, err := chat.mapRoom.GetRoomUsers(roomID)
	if err != nil {
		return err
	}

	// TODO: make error inside error too
	// send message to all user
	var userWg sync.WaitGroup
	// fmt.Println("User in rooms", userIDs)
	for _, userID := range userIDs {
		// fmt.Printf("\\-- User: %x\n", userID)
		userWg.Add(1)
		go func(userID string, wg *sync.WaitGroup) {
			// loop to all connection of user
			connIDs, err := chat.mapConn.GetConnectionByUser(userID)
			if err != nil {
				return
			}

			// send message to all conns of user
			var connWg sync.WaitGroup
			// fmt.Printf("user %s's connection %#v\n", userID, connIDs)
			for _, connID := range connIDs {
				// fmt.Printf("\\-- User %s: conn %s\n", userID, connID)
				connWg.Add(1)
				go func(connID string, wg *sync.WaitGroup) {
					fmt.Printf("[%s] --> %+v\n", connID, message)
					err := chat.send.SendMessage(connID, message)
					if err != nil {
						fmt.Printf("Error sending message: %s\n", err)
					}
					connWg.Done()
					// fmt.Println("Done ")
				}(connID, &connWg)
			}

			connWg.Wait()
			// end send message to all conn

			wg.Done()
		}(userID, &userWg)
	}
	userWg.Wait()
	// end send message to all user
	return nil
}

// OnConnect maange adding new connection, then return new ID to be used as reference when disconnect
func (chat *ChatDownstreamService) OnConnect(conn *chatsocket.SocketConnection) (connID string, err error) {
	connID, err = chat.mapConn.AddConnection(conn)
	fmt.Printf("[chat] user %s connected id = %s\n", conn.UserID, connID)
	return
}

// OnDisconnect should be called when client disconnect, connID should be obtained fron OnConnect
func (chat *ChatDownstreamService) OnDisconnect(connID string) error {
	err := chat.mapConn.RemoveConnection(connID)
	fmt.Printf("[chat] disconnected id = %s\n", connID)
	return err
}
