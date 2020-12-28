package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"fmt"
	"sync"
	"time"
)

// ChatService manages sending message and connection pool
type ChatService struct {
	mapRoomProxy repository.RoomUserRepository
	mapRoomUser  repository.RoomUserRepository
	send         repository.SendMessageRepository
	mapConn      repository.SocketConnectionRepository
	msgRepo      repository.MessageRepository
	notifService *NotificationService
}

// NewChatService create new instance of chat service
func NewChatService(roomProxyRepo repository.RoomUserRepository, roomUserRepo repository.RoomUserRepository, sender repository.SendMessageRepository, userConnRepo repository.SocketConnectionRepository, msgRepo repository.MessageRepository, notifService *NotificationService) *ChatService {
	return &ChatService{
		mapRoomProxy: roomProxyRepo,
		mapRoomUser:  roomUserRepo,
		send:         sender,
		mapConn:      userConnRepo,
		msgRepo:      msgRepo,
		notifService: notifService,
	}
}

// IsUserInRoom check whether user is in room
func (chat *ChatService) IsUserInRoom(userID string, roomID string) (bool, error) {
	rooms, err := chat.mapRoomProxy.GetUserRooms(userID)

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

	userIDs, err := chat.mapRoomProxy.GetRoomUsers(roomID)
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

// SendNotificationToRoom send notification to all users in the room whose last seen time is more than `thres`
// TODO refactor into another service
func (chat *ChatService) SendNotificationToRoom(roomID string, notification *model.Notification, thres time.Duration) error {
	userIDs, err := chat.mapRoomUser.GetRoomUsers(roomID)
	if err != nil {
		return err
	}

	allFCMTokens := make([]string, 0, len(userIDs)) // just pre allocate
	resultChan := make(chan []model.FCMToken, 1)

	for _, uid := range userIDs {
		go func(userID string) {
			// TODO handle error
			tokens, err := chat.notifService.GetUserTokens(userID)
			if err != nil {
				fmt.Printf("[send notif / get user tokens] for %s : error %s\n", userID, err.Error())
			}
			resultChan <- tokens
		}(uid)
	}

	for i := 0; i < len(userIDs); i++ {
		for _, tok := range <-resultChan {
			online := chat.notifService.GetOnlineStatus(tok.Token)
			if online {
				fmt.Print("ignore device", tok.Token[:10], "since it's online")
				continue
			}
			lastSeen := chat.notifService.GetLastSeenTime(tok.Token)
			if lastSeen.IsZero() || time.Now().Sub(lastSeen) > thres {
				allFCMTokens = append(allFCMTokens, tok.Token)
			} else {
				fmt.Print("ignore device", tok.Token[:10], "since it's seen less than", thres, "ago")
			}
		}
	}

	// TODO handle later
	if len(allFCMTokens) > 500 {
		return fmt.Errorf("too many device to send")
	}

	success, err := chat.notifService.SendNotifications(allFCMTokens, notification)
	fmt.Printf("[notification] successfully sent %d of %d notifications\n", success, len(allFCMTokens))
	return err
}

// OnConnect maange adding new connection, then return new ID to be used as reference when disconnect
func (chat *ChatService) OnConnect(conn *chatsocket.Connection) (connID string, err error) {
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
