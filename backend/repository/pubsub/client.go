package pubsub

import (
	"backendSenior/repository"
	"bytes"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type subscription struct {
	conn *connection
	room string
}

type connection struct {
	// The websocket connection.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	send       chan []byte
	clientsMtx sync.Mutex
}

func (s *subscription) readPump() {
	c := s.conn
	defer func() {
		//Unregister
		H.unregister <- *s
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		//Reading incoming message...
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		m := message{s.room, msg}
		H.broadcast <- m
	}
}
func (s *subscription) writePump() {
	c := s.conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		//Listerning message when it comes will write it into writer and then send it to the client
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.writeDB(websocket.TextMessage, message, s.room); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
func ServeWs(context *gin.Context) {
	w := context.Writer
	r := context.Request
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	//Get room's id from client...
	queryValues := r.URL.Query()
	roomId := queryValues.Get("roomId")

	//TODO :: if it a room in database ??

	c := &connection{send: make(chan []byte, 256), ws: ws}
	s := subscription{c, roomId}
	H.register <- s
	go s.writePump()
	go s.readPump()
}

func (c *connection) writeDB(mt int, payload []byte, room string) error {
	// Lock session to write to DB
	//c.clientsMtx.Lock()
	// Fix :: ?? May cuase memory crash
	err := repository.AddMessageDB(payload, room, "")
	if err != nil {
		log.Println("error add Message to DB")
	}
	//c.clientsMtx.Unlock()
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}
