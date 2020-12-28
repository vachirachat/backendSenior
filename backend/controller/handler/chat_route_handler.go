package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/exception"
	"backendSenior/domain/model/chatsocket/message_types"
	"backendSenior/domain/service"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// ChatRouteHandler is handler for real time chat (websocket)
type ChatRouteHandler struct {
	chatService *service.ChatService
	proxyMw     *auth.ProxyMiddleware
	roomService *service.RoomService // for mapping
}

// NewChatRouteHandler create new `ChatRouteHandler`
func NewChatRouteHandler(chatService *service.ChatService, proxyMw *auth.ProxyMiddleware, roomSvc *service.RoomService) *ChatRouteHandler {
	return &ChatRouteHandler{
		chatService: chatService,
		proxyMw:     proxyMw,
		roomService: roomSvc,
	}
}

// client abstraction
type client struct {
	conn        *websocket.Conn
	chatService *service.ChatService // reference chat service to call
	roomService *service.RoomService
	connID      string
	proxyID     string
}

//Mount make the handler handle request from specfied routerGroup
func (handler *ChatRouteHandler) Mount(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/ws", handler.proxyMw.AuthRequired(), func(context *gin.Context) {
		// fmt.Println("new connection!")
		w := context.Writer
		r := context.Request

		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(req *http.Request) bool {
				return true
			},
		}

		clientID := context.GetString(auth.UserIdField)

		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		var conn = &chatsocket.Connection{
			Conn:   wsConn,
			UserID: clientID,
		}

		id, err := handler.chatService.OnConnect(conn)

		clnt := client{
			conn:        wsConn,
			chatService: handler.chatService,
			roomService: handler.roomService,
			connID:      id,
			proxyID:     clientID,
		}

		go clnt.readPump()
	})
}

// TODO: some how move this to connection pool so it's centralized
// readPump is for reading message and call handler
// write pump code is in connection pool
// for more information about read/writePump, see https://github.com/gorilla/websocket/tree/master/examples/chat
func (c *client) readPump() {
	defer func() {
		c.chatService.OnDisconnect(c.connID)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		// fmt.Printf("[chat] <-- %s\n", message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// handle message here
		fmt.Printf("[%s] <-- %s\n", c.connID, message)

		var rawMessage chatsocket.RawMessage
		err = json.Unmarshal(message, &rawMessage)
		if err != nil {
			c.chatService.SendMessageToConnection(c.connID, chatsocket.Message{
				Type: message_types.Error,
				Payload: exception.Event{
					Reason: "Bad socket message structure",
				},
			})
			continue
		}

		switch rawMessage.Type {
		case message_types.Chat:
			var msg model.Message
			err = json.Unmarshal(rawMessage.Payload, &msg)
			if err != nil {
				c.chatService.SendMessageToConnection(c.connID, chatsocket.Message{
					Type: message_types.Error,
					Payload: exception.Event{
						Reason: "bad message payload format",
						Data:   err.Error(),
					},
				})
				continue
			}
			if ok, err := c.chatService.IsUserInRoom(c.proxyID, msg.RoomID.Hex()); err != nil {
				c.chatService.SendMessageToConnection(c.connID, chatsocket.Message{
					Type: message_types.Error,
					Payload: exception.Event{
						Reason: "unable to check room",
						Data:   err.Error(),
					},
				})
				continue
			} else if !ok {
				c.chatService.SendMessageToConnection(c.connID, chatsocket.Message{
					Type: message_types.Error,
					Payload: exception.Event{
						Reason: "unauthorized to send message to the room",
					},
				})
				continue
			}
			// Saving messag
			msg.TimeStamp = time.Now()
			if msg.UserID == "" {
				fmt.Println("Bad Message, No User ID (proxy must fill it)")
				continue
			}

			messageID, err := c.chatService.SaveMessage(msg)
			if err != nil {
				fmt.Printf("error saving message %s\n", err.Error())
				continue
			}
			msg.MessageID = bson.ObjectIdHex(messageID)

			err = c.chatService.BroadcastMessageToRoom(msg.RoomID.Hex(), chatsocket.Message{
				Type:    message_types.Chat,
				Payload: msg,
			})
			if err != nil {
				fmt.Printf("Error bcasting message: %s\n", err.Error())
			}

			c.chatService.SendNotificationToRoomExceptUser(msg.RoomID.Hex(), msg.UserID.Hex(), &model.Notification{
				// Title: "New Message in room " + msg.RoomID.Hex(),
				// Body:  fmt.Sprintf("[%s]: %s", msg.UserID.Hex(), msg.Data),
				Data: map[string]string{
					"roomId":    msg.RoomID.Hex(),
					"msgId":     msg.MessageID.Hex(),
					"userId":    msg.UserID.Hex(),
					"timestamp": msg.TimeStamp.Format("2006-01-02T15:04:05Z"),
				},
			}, 1*time.Second)

		default:
			fmt.Printf("INFO: unrecognized message\n==\n%s\n==\n", message)

		}

	}
}
