package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// ChatService manages sending message and connection pool
type ChatService struct {
	mapRoomProxy repository.RoomProxyRepository
	mapRoomUser  repository.RoomUserRepository
	send         repository.SendMessageRepository
	mapConn      repository.SocketConnectionRepository
	msgRepo      repository.MessageRepository
	notifService *NotificationService
	log          *log.Logger
}

// NewChatService create new instance of chat service
func NewChatService(roomProxyRepo repository.RoomProxyRepository, roomUserRepo repository.RoomUserRepository, sender repository.SendMessageRepository, userConnRepo repository.SocketConnectionRepository, msgRepo repository.MessageRepository, notifService *NotificationService) *ChatService {
	return &ChatService{
		mapRoomProxy: roomProxyRepo,
		mapRoomUser:  roomUserRepo,
		send:         sender,
		mapConn:      userConnRepo,
		msgRepo:      msgRepo,
		notifService: notifService,
		log:          log.New(os.Stdout, "ChatService:", log.Ldate|log.Lshortfile),
	}
}

// IsProxyInRoom check whether proxy is in room and allowed to send message
func (chat *ChatService) IsProxyInRoom(proxyID string, roomID string) (bool, error) {
	rooms, err := chat.mapRoomProxy.GetProxyRooms(proxyID)

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
func (chat *ChatService) BroadcastMessageToRoom(roomID string, data interface{}) error {

	userIDs, err := chat.mapRoomProxy.GetRoomProxies(roomID)
	if err != nil {
		return fmt.Errorf("getting room's users: %s", err)
	}

	chat.log.Println("room", roomID, "has users", userIDs)

	var allWg sync.WaitGroup

	for _, uid := range userIDs {
		connIDs, err := chat.mapConn.GetConnectionByUser(uid)
		//chat.log.Println("user", uid, "has connections", connIDs)

		if err != nil {
			chat.log.Println("error getting user connection", err)
			return err
		}
		for _, connID := range connIDs {
			allWg.Add(1)
			go func(connID string, wg *sync.WaitGroup) {
				defer wg.Done()
				chat.log.Println("sending message to conn", connID)
				// loop to all connection of user
				err := chat.send.SendMessage(connID, data)
				if err != nil {
					fmt.Println("Error sending message", err)
				}
			}(connID, &allWg)
		}
	}

	allWg.Wait()
	return nil
}

// SendNotificationToRoomExceptUser send notification to all users in the room whose last seen time is more than `thres`
// It exclude userID (sender) from receiving message
// TODO refactor into another service
func (chat *ChatService) SendNotificationToRoomExceptUser(roomID string, userID string, notification *model.Notification, thres time.Duration) error {
	userIDs, err := chat.mapRoomUser.GetRoomUsers(roomID)
	if err != nil {
		return err
	}

	allFCMTokens := make([]string, 0, len(userIDs)) // just pre allocate
	resultChan := make(chan []model.FCMToken, 1)

	cnt := 0
	for _, uid := range userIDs {
		if uid == userID {
			fmt.Println("excluded user", uid, "as it's sender")
			continue
		}
		cnt += 1
		go func(userID string) {
			// TODO handle error
			tokens, err := chat.notifService.GetUserTokens(userID)
			if err != nil {
				fmt.Printf("[send notif / get user tokens] for %s : error %s\n", userID, err.Error())
			}
			resultChan <- tokens
		}(uid)
	}

	for i := 0; i < cnt; i++ {
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
	rooms, err := chat.mapRoomProxy.GetProxyRooms(conn.UserID)
	chat.log.Printf("client %s connected, connection id %s\n", conn.UserID, conn.ConnID)

	for _, roomID := range rooms {
		go chat.BroadcastMessageToRoom(roomID, chatsocket.InvalidateRoomMasterMessage(roomID))
	}
	return
}

// OnDisconnect should be called when client disconnect, connID should be obtained fron OnConnect
func (chat *ChatService) OnDisconnect(conn *chatsocket.Connection) error {
	err := chat.mapConn.RemoveConnection(conn.ConnID)
	rooms, err := chat.mapRoomProxy.GetProxyRooms(conn.UserID)
	chat.log.Printf("client %s dis-connected, connection id %s\n", conn.UserID, conn.ConnID)

	for _, roomID := range rooms {
		go chat.BroadcastMessageToRoom(roomID, chatsocket.InvalidateRoomMasterMessage(roomID))
	}

	return err
}
