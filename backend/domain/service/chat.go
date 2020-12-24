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

// IsUserInRoom check whether user is in room
func (chat *ChatService) IsUserInRoom(userID string, roomID string) (bool, error) {
	rooms, err := chat.mapRoom.GetUserRooms(userID)

	if err != nil {
		return false, err
	}
	for _, r := range rooms {
		if r == roomID {
			return true, nil
		}
	}
	return false, nil
}

// SaveMessage save speicified message to repository, returning the ID of message
func (chat *ChatService) SaveMessage(message model.Message) (string, error) {
	id, err := chat.msgRepo.AddMessage(message)
	return id, err
}

// SendMessageToConnection send message to specific connection, data will be marshalled
func (chat *ChatService) SendMessageToConnection(connID string, message interface{}) error {
	return chat.send.SendMessage(connID, message)
}

// BroadcastMessageToRoom send message to socket of all users in the room
// []byte will be sent as is, but other value will be marshalled
// TODO: this is currently broadcast to all
func (chat *ChatService) BroadcastMessageToRoom(roomID string, data interface{}) error {

	userIDs, err := chat.mapRoom.GetRoomUsers(roomID)
	if err != nil {
		return fmt.Errorf("getting room's users: %s", err)
	}

	fmt.Println("room", roomID, "has users", userIDs)

	var allWg sync.WaitGroup

	for _, uid := range userIDs {
		allWg.Add(1)

		go func(uid string, wg *sync.WaitGroup) {
			defer wg.Done()

			connIDs, err := chat.mapConn.GetConnectionByUser(uid)
			if err != nil {
				fmt.Printf("error getting user connections: %s\n", err.Error())
				return
			}
			// TODO: make error inside error too

			var userWg sync.WaitGroup
			// fmt.Println("User in rooms", userIDs)
			for _, connID := range connIDs {
				// fmt.Printf("\\-- User: %x\n", userID)
				userWg.Add(1)
				go func(connID string, wg *sync.WaitGroup) {
					defer wg.Done()
					// loop to all connection of user
					err := chat.send.SendMessage(connID, data)
					if err != nil {
						fmt.Println("Error sending message", err)
					}
				}(connID, &userWg)
			}

			userWg.Wait() // wait all user done

		}(uid, &allWg)
	}

	allWg.Wait()

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
