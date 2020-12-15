package route

import (
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
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
}

// NewChatRouteHandler create new `ChatRouteHandler`
func NewChatRouteHandler(chatService *service.ChatService) *ChatRouteHandler {
	return &ChatRouteHandler{
		chatService: chatService,
	}
}

// client abstraction
type client struct {
	conn        *websocket.Conn
	chatService *service.ChatService // reference chat service to call
	id          string
}

//Mount make the handler handle request from specfied routerGroup
func (handler *ChatRouteHandler) Mount(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/ws", func(context *gin.Context) {
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

		// Proxy use no auth ?
		// TODO: distinguish by ID, but for now, just bcast to ALL
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		var conn = &chatsocket.SocketConnection{
			Conn:   wsConn,
			UserID: "foo",
		}

		id, err := handler.chatService.OnConnect(conn)

		clnt := client{
			conn:        wsConn,
			chatService: handler.chatService,
			id:          id,
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
		c.chatService.OnDisconnect(c.id)
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
		fmt.Printf("[%s] <-- %s\n", c.id, message)
		var msg model.Message
		json.Unmarshal(message, &msg)

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

		// TODO: this is broadbast to ALL proxy for now
		err = c.chatService.BroadcastMessageToRoom(msg.RoomID.Hex(), msg)
		if err != nil {
			fmt.Printf("Error bcasting message: %s\n", err.Error())
		}
	}
}
