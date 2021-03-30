package route

import (
	"backendSenior/domain/model"
	"backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/exception"
	"backendSenior/domain/model/chatsocket/message_types"
	"common/ws"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"proxySenior/controller/middleware"
	"proxySenior/domain/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/websocket"
)

type ChatRouteHandler struct {
	upstream       *service.ChatUpstreamService
	downstream     *service.ChatDownstreamService
	authMiddleware *middleware.AuthMiddleware
	//key            *key_service.KeyService
	encryption *service.EncryptionService
}

func NewChatRouteHandler(upstream *service.ChatUpstreamService, downstream *service.ChatDownstreamService, authMw *middleware.AuthMiddleware, enc *service.EncryptionService) *ChatRouteHandler {
	return &ChatRouteHandler{
		upstream:       upstream,
		downstream:     downstream,
		authMiddleware: authMw,
		encryption:     enc,
	}
}

// client abstraction
type client struct {
	chatsocket *chatsocket.Connection
	handlerRef *ChatRouteHandler
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
	rawConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	wsConn := ws.FromConnection(rawConn)
	wsConn.StartLoop()
	var chatSocket = &chatsocket.Connection{
		Conn:   wsConn,
		UserID: userID,
	}

	id, err := handler.downstream.OnConnect(chatSocket)
	clnt := client{
		chatsocket: chatSocket,
		handlerRef: handler,
	}

	clnt.readLoop()
	chatSocket.Conn.Observable().DoOnCompleted(func() {
		_ = handler.downstream.OnDisconnect(id)

	})
}

// TODO: duplicate code
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

// readLoop
func (c *client) readLoop() {
	conn := c.chatsocket.Conn
	userID := c.chatsocket.UserID
	_ = userID // TODO: use to check permission
	<-conn.Observable().DoOnNext(func(i interface{}) {
		inMessage := i.([]byte)
		var rawMessage chatsocket.RawMessage

		if err := json.Unmarshal(inMessage, &rawMessage); err != nil {
			conn.SendJSON(wsErrorMessage("bad socket message structure"))
			return
		}

		switch rawMessage.Type {
		case message_types.Chat:
			// handle message here
			var msg model.Message
			if err := json.Unmarshal(rawMessage.Payload, &msg); err != nil {
				fmt.Println("bad message payload format")
				conn.SendJSON(wsErrorMessage("bad message payload format", err))
				return
			}

			if ok, err := c.handlerRef.downstream.IsUserInRoom(userID, msg.RoomID.Hex()); err != nil {
				fmt.Println("unable to check room")
				conn.SendJSON(wsErrorMessage("unable to check room", err))
				return
			} else if !ok {
				conn.SendJSON(wsErrorMessage("unauthorized"))
				return
			}

			now := time.Now()
			if err := c.handlerRef.encryption.EncryptController(&msg); err != nil {
				conn.SendJSON(wsErrorMessage("encryption error", err))
				return
			}
			msg.TimeStamp = now
			msg.UserID = bson.ObjectIdHex(userID)
			if err := c.handlerRef.upstream.SendMessage(msg); err != nil {
				fmt.Println("error sending", err)
				conn.SendJSON(wsErrorMessage("error sending message to controller", err))
				return
			}
			// TODO: add send success here
		default:
			conn.SendJSON(wsErrorMessage("unsupported message format"))
			fmt.Printf("INFO: unrecognized message\n==\n%s\n==\n", inMessage)
		}
	})
}
