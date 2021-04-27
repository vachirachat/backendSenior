package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/service"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ConnStateRouteHandler struct {
	notifService *service.NotificationService
	authMw       *auth.JWTMiddleware
}

func NewConnStateRouteHandler(notifService *service.NotificationService, authMw *auth.JWTMiddleware) *ConnStateRouteHandler {
	return &ConnStateRouteHandler{
		notifService: notifService,
		authMw:       authMw,
	}
}

func (h *ConnStateRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("/ws", h.authMw.AuthRequired(), h.handleWs)
}

type simpleClient struct {
	conn               *websocket.Conn
	closeChan          chan struct{}
	updateLastSeenChan chan struct{}
	sendChan           chan []byte
}

func (h *ConnStateRouteHandler) handleWs(c *gin.Context) {
	// fmt.Println("new connection!")
	w := c.Writer
	r := c.Request

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(req *http.Request) bool {
			return true
		},
	}

	userID := c.GetString(auth.UserIdField)

	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		v := c.Request.Header["authorization"]
		if len(v) != 0 {
			authHeader = v[0]
		}
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 3 {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "bad auth header"})
		return
	}

	fcmToken := parts[2]
	tok, err := h.notifService.GetTokenByID(fcmToken)
	if err != nil {
		fmt.Println("[conn state] get token detail error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	if tok.UserID.Hex() != userID {
		c.JSON(http.StatusForbidden, gin.H{"status": "not your token"})
		return
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[conn state] error upgrading", err)
		return
	}
	fmt.Println("clnt", fcmToken[:10], "connected", time.Now())
	h.notifService.SetOnlineStatus(fcmToken, true)
	h.notifService.SetLastSeenTime(fcmToken, time.Now())

	clnt := &simpleClient{
		conn:               wsConn,
		closeChan:          make(chan struct{}),
		updateLastSeenChan: make(chan struct{}),
		sendChan:           make(chan []byte),
	}

	// simple websocket that do nothing
	go clnt.readPump()
	go clnt.writePump()

	go func() {
		for {
			select {
			case <-clnt.updateLastSeenChan:
				fmt.Println("clnt", fcmToken[:10], "last seen", time.Now())
				h.notifService.SetLastSeenTime(fcmToken, time.Now())
				h.notifService.SetOnlineStatus(fcmToken, true)
			case <-clnt.closeChan:
				fmt.Println("clnt", fcmToken[:10], "disconnected", time.Now())
				h.notifService.SetOnlineStatus(fcmToken, false)
				return
			}
		}
	}()
}

func (c *simpleClient) readPump() {
	defer func() {
		close(c.closeChan)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.updateLastSeenChan <- struct{}{}
		return nil
	})
	for {
		_, _, err := c.conn.ReadMessage()
		// fmt.Printf("[chat] <-- %s\n", message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func (c *simpleClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.conn.Close()
		ticker.Stop()
	}()

	for {
		select {
		case msg, ok := <-c.sendChan:
			// hub closed connection
			if !ok {
				c.conn.WriteMessage(websocket.TextMessage, []byte{})
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("error writing: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
