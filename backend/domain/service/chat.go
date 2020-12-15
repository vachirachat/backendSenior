package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"fmt"
	"sync"
)

// ChatService manages sending message and connection pool
type ChatService struct {
	mapRoom repository.RoomUserRepository
	send    repository.SendMessageRepository
	mapConn repository.SocketConnectionRepository
	msgRepo repository.MessageRepository
}

// NewChatService create new instance of chat service
func NewChatService(roomUserRepo repository.RoomUserRepository, sender repository.SendMessageRepository, userConnRepo repository.SocketConnectionRepository, msgRepo repository.MessageRepository) *ChatService {
	return &ChatService{
		mapRoom: roomUserRepo,
		send:    sender,
		mapConn: userConnRepo,
		msgRepo: msgRepo,
	}
}

// SaveMessage save speicified message to repository, returning the ID of message
func (chat *ChatService) SaveMessage(message model.Message) (string, error) {
	id, err := chat.msgRepo.AddMessage(message)
	return id, err
}

// BroadcastMessageToRoom send message to socket of all users in the room
// []byte will be sent as is, but other value will be marshalled
// TODO: this is currently broadcast to all
func (chat *ChatService) BroadcastMessageToRoom(roomID string, data interface{}) error {
	// TODO: for now user id is always foo
	connIDs, err := chat.mapConn.GetConnectionByUser("foo")
	if err != nil {
		return err
	}
	// TODO: make error inside error too
	// send message to all user
	var connWg sync.WaitGroup
	// fmt.Println("User in rooms", userIDs)
	for _, connID := range connIDs {
		// fmt.Printf("\\-- User: %x\n", userID)
		connWg.Add(1)
		go func(connID string, wg *sync.WaitGroup) {
			// loop to all connection of user
			err := chat.send.SendMessage(connID, data)
			if err != nil {
				fmt.Println("Error sending message", err)
			}
			wg.Done()
		}(connID, &connWg)
	}
	connWg.Wait()
	// end send message to all user
	return nil
}

// OnConnect maange adding new connection, then return new ID to be used as reference when disconnect
func (chat *ChatService) OnConnect(conn *chatsocket.SocketConnection) (connID string, err error) {
	connID, err = chat.mapConn.AddConnection(conn)
	fmt.Printf("[chat] user %s connected id = %s\n", conn.UserID, connID)
	return
}

// OnDisconnect should be called when client disconnect, connID should be obtained fron OnConnect
func (chat *ChatService) OnDisconnect(connID string) error {
	err := chat.mapConn.RemoveConnection(connID)
	fmt.Printf("[chat] disconnected id = %s\n", connID)
	return err
}
