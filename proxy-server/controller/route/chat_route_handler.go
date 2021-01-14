package route

import (
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/exception"
	"backendSenior/domain/model/chatsocket/message_types"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"proxySenior/controller/middleware"
	"proxySenior/domain/encryption"
	model_proxy "proxySenior/domain/model"
	"proxySenior/domain/service"
	"proxySenior/domain/service/key_service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/websocket"
)

type ChatRouteHandler struct {
	upstream       *service.ChatUpstreamService
	downstream     *service.ChatDownstreamService
	authMiddleware *middleware.AuthMiddleware
	key            *key_service.KeyService
}

func NewChatRouteHandler(upstream *service.ChatUpstreamService, downstream *service.ChatDownstreamService, authMw *middleware.AuthMiddleware, key *key_service.KeyService) *ChatRouteHandler {
	return &ChatRouteHandler{
		upstream:       upstream,
		downstream:     downstream,
		authMiddleware: authMw,
		key:            key,
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
	handlerRef *ChatRouteHandler
	connID     string
	userID     string
}

//Mount make the handler handle request from specfied routerGroup
func (handler *ChatRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("/ws", handler.authMiddleware.AuthRequired(), handler.websocketHandler)
}

func (handler *ChatRouteHandler) websocketHandler(context *gin.Context) {
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

	userID := context.GetString(middleware.UserIdField)

	// Proxy use no auth ?
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var conn = &chatsocket.Connection{
		Conn:   wsConn,
		UserID: userID,
	}

	id, err := handler.downstream.OnConnect(conn)

	clnt := client{
		conn:       wsConn,
		handlerRef: handler,
		connID:     id,
		userID:     userID,
	}
	go clnt.readPump()
}

// TODO: some how move this to connection pool so it's centralized
// readPump is for reading message and call handler
// write pump code is in connection pool
// for more information about read/writePump, see https://github.com/gorilla/websocket/tree/master/examples/chat
func (c *client) readPump() {
	defer func() {
		c.handlerRef.downstream.OnDisconnect(c.connID)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, inMessage, err := c.conn.ReadMessage()
		// fmt.Printf("[chat] <-- %s\n", message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		inMessage = bytes.TrimSpace(bytes.Replace(inMessage, newline, space, -1))
		var rawMessage chatsocket.RawMessage
		err = json.Unmarshal(inMessage, &rawMessage)
		if err != nil {
			c.handlerRef.downstream.SendMessageToConnection(c.connID, chatsocket.Message{
				Type: message_types.Error,
				Payload: exception.Event{
					Reason: "Bad socket message structure",
				},
			})
			continue
		}

		fmt.Printf("[%s] <-- %s\n", c.connID, inMessage)
		switch rawMessage.Type {
		case message_types.Chat:
			// handle message here
			var msg model.Message
			err = json.Unmarshal(rawMessage.Payload, &msg)

			if err != nil {
				fmt.Println("bad message payload format")
				c.handlerRef.downstream.SendMessageToConnection(c.connID, chatsocket.Message{
					Type: message_types.Error,
					Payload: exception.Event{
						Reason: "bad message payload format",
						Data:   err.Error(),
					},
				})
				continue
			}

			if ok, err := c.handlerRef.downstream.IsUserInRoom(c.userID, msg.RoomID.Hex()); err != nil {
				fmt.Println("unable to check room")
				c.handlerRef.downstream.SendMessageToConnection(c.connID, chatsocket.Message{
					Type: message_types.Error,
					Payload: exception.Event{
						Reason: "unable to check room",
						Data:   err.Error(),
					},
				})
				continue
			} else if !ok {
				fmt.Println("unauthorized", c.userID, "not in room", msg.RoomID.Hex())
				c.handlerRef.downstream.SendMessageToConnection(c.connID, chatsocket.Message{
					Type: message_types.Error,
					Payload: exception.Event{
						Reason: "unauthorized to send message to the room",
					},
				})
				continue
			}

			keys, err := c.handlerRef.getKeyFromRoom(msg.RoomID.Hex())
			if err != nil {
				fmt.Println("can't get key for room:", msg.RoomID.Hex())
				c.handlerRef.downstream.SendMessageToConnection(c.connID, chatsocket.Message{
					Type: message_types.Error,
					Payload: exception.Event{
						Reason: "can't get key for room: " + msg.RoomID.Hex(),
					},
				})
				continue
			}

			now := time.Now()
			key := keyFor(keys, now)
			encrypted, err := encryption.AESEncrypt([]byte(msg.Data), key)
			if err != nil {
				fmt.Println("encryption error:", err)
				c.handlerRef.downstream.SendMessageToConnection(c.connID, chatsocket.Message{
					Type: message_types.Error,
					Payload: exception.Event{
						Reason: "encryption error: " + err.Error(),
					},
				})
				continue
			}
			encrypted = encryption.B64Encode(encrypted)
			msg.Data = string(encrypted)

			msg.TimeStamp = now
			msg.UserID = bson.ObjectIdHex(c.userID)
			err = c.handlerRef.upstream.SendMessage(msg)
			if err != nil {
				fmt.Println("error sending")
				c.handlerRef.downstream.SendMessageToConnection(c.connID, chatsocket.Message{
					Type: message_types.Error,
					Payload: exception.Event{
						Reason: "error sending message to controller",
						Data:   err.Error(),
					},
				})
				continue
			}
		default:
			fmt.Printf("INFO: unrecognized message\n==\n%s\n==\n", inMessage)
		}
	}
}

// TODO: this is duplicate code as in message_handler.go
//getKeyFromRoom determine where to get key and get the key
func (h *ChatRouteHandler) getKeyFromRoom(roomID string) ([]model_proxy.KeyRecord, error) {
	local, err := h.key.IsLocal(roomID)
	if err != nil {
		return nil, fmt.Errorf("error deftermining locality ok key %v", err)
	}

	var keys []model_proxy.KeyRecord
	if local {
		fmt.Println("[message] use LOCAL key for", roomID)
		keys, err = h.key.GetKeyLocal(roomID)
		if err != nil {
			return nil, fmt.Errorf("error getting key locally %v", err)
		}
	} else {
		fmt.Println("[message] use REMOTE key for room", roomID)
		keys, err = h.key.GetKeyRemote(roomID)
		if err != nil {
			return nil, fmt.Errorf("error getting key remotely %v", err)
		}
	}

	return keys, nil
}
