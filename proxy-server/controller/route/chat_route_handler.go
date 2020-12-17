package route

import (
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"proxySenior/domain/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/websocket"
)

type ChatRouteHandler struct {
	upstream   *service.ChatUpstreamService
	downstream *service.ChatDownstreamService
}

func NewChatRouteHandler(upstream *service.ChatUpstreamService, downstream *service.ChatDownstreamService) *ChatRouteHandler {
	return &ChatRouteHandler{
		upstream:   upstream,
		downstream: downstream,
	}
}

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

// client abstraction
type client struct {
	conn       *websocket.Conn
	upstream   *service.ChatUpstreamService
	downstream *service.ChatDownstreamService // reference chat service to call
	connID     string
	userID     string
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

		userID := context.Query("userID")
		if userID == "" {
			context.JSON(http.StatusBadRequest, "Must Specify userID to connect")
		}
		// Proxy use no auth ?
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		var conn = &chatsocket.SocketConnection{
			Conn:   wsConn,
			UserID: userID,
		}

		id, err := handler.downstream.OnConnect(conn)

		clnt := client{
			conn:       wsConn,
			upstream:   handler.upstream,
			downstream: handler.downstream,
			connID:     id,
			userID:     userID,
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
		c.downstream.OnDisconnect(c.connID)
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
		var msg model.Message
		json.Unmarshal(message, &msg)

		// Saving messag
		msg.TimeStamp = time.Now()
		msg.UserID = bson.ObjectIdHex(c.userID)

		err = c.upstream.SendMessage(msg)
		if err != nil {
			fmt.Println("Error sending to upstream:", err)
			continue
		}

		// messageID, err := c.downstream.SaveMessage(msg)
		// if err != nil {
		// 	fmt.Printf("error saving message %s\n", err.Error())
		// 	continue
		// }
		// msg.MessageID = bson.ObjectIdHex(messageID)

		// // TODO: this is broadbast to ALL proxy for now
		// err = c.downstream.BroadcastMessageToRoom(msg.RoomID.Hex(), msg)
		// if err != nil {
		// 	fmt.Printf("Error bcasting message: %s\n", err.Error())
		// }
	}
}
