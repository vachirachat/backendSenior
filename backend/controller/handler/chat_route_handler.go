package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/exception"
	"backendSenior/domain/model/chatsocket/message_types"
	"backendSenior/domain/service"
	"common/ws"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
	maxMessageSize = 4096
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// ChatRouteHandler is handler for real time chat (websocket)
type ChatRouteHandler struct {
	chatService *service.ChatService
	proxyMw     *auth.ProxyMiddleware
	keyEx       *service.KeyExchangeService
	roomService *service.RoomService // for mapping
	log         *log.Logger
}

// NewChatRouteHandler create new `ChatRouteHandler`
func NewChatRouteHandler(chatService *service.ChatService, proxyMw *auth.ProxyMiddleware, roomSvc *service.RoomService, keyEx *service.KeyExchangeService) *ChatRouteHandler {
	return &ChatRouteHandler{
		chatService: chatService,
		proxyMw:     proxyMw,
		roomService: roomSvc,
		keyEx:       keyEx,
		log:         log.New(os.Stdout, "[chat-route-handler]", log.Ldate|log.Lshortfile),
	}
}

// client is like chat socket, but added handler field to allow access to handler related function
type client struct {
	chatsocket.Connection
	handler *ChatRouteHandler // reference to handler to call
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

		proxyID := context.GetString(auth.UserIdField)

		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		var conn = &chatsocket.Connection{
			Conn:   ws.FromConnection(wsConn),
			UserID: proxyID,
		}

		_, err = handler.chatService.OnConnect(conn)
		handler.keyEx.SetOnline(proxyID, true)

		clnt := client{
			Connection: chatsocket.Connection{},
			handler:    handler,
		}

		clnt.handleMessage()
		conn.Conn.Observable().DoOnCompleted(func() {
			handler.chatService.OnDisconnect(conn)
			// here we assume that proxy has ONLY ONE connection
			handler.keyEx.SetOnline(proxyID, false)

		})

	})
}

func wsErrorMessage(reason string, data ...interface{}) chatsocket.Message {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	return chatsocket.Message{
		Type: message_types.Error,
		Payload: exception.Event{
			Reason: reason,
			Data:   d,
		},
	}
}

// TODO: some how move this to connection pool so it's centralized
// handleMessage is for reading message and call handler
// write pump code is in connection pool
// for more information about read/writePump, see https://github.com/gorilla/websocket/tree/master/examples/chat
func (c *client) handleMessage() {
	wsConn := c.Conn
	proxyID := c.UserID
	handler := c.handler

	wsConn.Observable().DoOnNext(func(i interface{}) {
		message := i.([]byte)

		var rawMessage chatsocket.RawMessage
		if err := json.Unmarshal(message, &rawMessage); err != nil {
			wsConn.SendJSON(wsErrorMessage("bad socket message structure", err))
			return
		}

		switch rawMessage.Type {
		case message_types.Chat:
			var msg model.Message
			if err := json.Unmarshal(rawMessage.Payload, &msg); err != nil {
				wsConn.SendJSON(wsErrorMessage("bad message payload format", err))
				return
			}

			if ok, err := handler.chatService.IsProxyInRoom(proxyID, msg.RoomID.Hex()); err != nil {
				wsConn.SendJSON(wsErrorMessage("can't check room", err))
				return
			} else if !ok {
				wsConn.SendJSON(wsErrorMessage("unauthorized"))
				return
			}

			// Saving messag
			msg.TimeStamp = time.Now()
			if msg.UserID == "" {
				wsConn.SendJSON(wsErrorMessage("bad message: missing user ID"))
				return
			}

			messageID, err := handler.chatService.SaveMessage(msg)
			if err != nil {
				fmt.Printf("error saving message %s\n", err.Error())
				wsConn.SendJSON(wsErrorMessage("send failed: error saving message"))
				return
			}
			msg.MessageID = bson.ObjectIdHex(messageID)
			if err = handler.chatService.BroadcastMessageToRoom(msg.RoomID.Hex(), chatsocket.Message{
				Type:    message_types.Chat,
				Payload: msg,
			}); err != nil {
				fmt.Printf("Error bcasting message: %s\n", err.Error())
			}

			handler.chatService.SendNotificationToRoomExceptUser(msg.RoomID.Hex(), msg.UserID.Hex(), &model.Notification{
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

	})

}
