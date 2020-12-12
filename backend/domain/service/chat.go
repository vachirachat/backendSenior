package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"fmt"
	"sync"

	"github.com/globalsign/mgo/bson"
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

// SaveMessage save speicified message to repository, returning the objectID of message
func (chat *ChatService) SaveMessage(message model.Message) (bson.ObjectId, error) {
	idHex, err := chat.msgRepo.AddMessage(message)
	return bson.ObjectIdHex(idHex), err
}

// TODO: in the future there should be broadcast event etc.
// BroadcastToRoom send message to socket of all users in the room
// []byte will be sent as is, but other value will be marshalled
func (chat *ChatService) BroadcastMessageToRoom(roomID string, data interface{}) error {
	userIDs, err := chat.mapRoom.GetRoomUsers(roomID)
	if err != nil {
		return err
	}

	// TODO: make error inside error too
	// send message to all user
	var userWg sync.WaitGroup
	fmt.Println("User in rooms", userIDs)
	for _, userID := range userIDs {
		fmt.Printf("\\-- User: %x\n", userID)
		userWg.Add(1)
		go func(userID string, wg *sync.WaitGroup) {
			// loop to all connection of user
			connIDs, err := chat.mapConn.GetConnectionByUser(userID)
			if err != nil {
				return
			}

			// send message to all conns of user
			var connWg sync.WaitGroup
			fmt.Printf("user %s's connection %#v\n", userID, connIDs)
			for _, connID := range connIDs {
				fmt.Printf("\\-- User %s: conn %s\n", userID, connID)
				connWg.Add(1)
				go func(connID string, wg *sync.WaitGroup) {
					err := chat.send.SendMessage(connID, data)
					if err != nil {
						fmt.Printf("Error sending message: %s\n", err)
					}
					connWg.Done()
					fmt.Println("Done ")
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
func (chat *ChatService) OnConnect(conn *chatsocket.SocketConnection) (connID string, err error) {
	connID, err = chat.mapConn.AddConnection(conn)
	return
}

// OnDisconnect should be called when client disconnect, connID should be obtained fron OnConnect
func (chat *ChatService) OnDisconnect(connID string) error {
	err := chat.mapConn.RemoveConnection(connID)
	return err
}
