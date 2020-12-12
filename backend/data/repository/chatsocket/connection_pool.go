package chatsocket

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model/chatsocket"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

// ConnectionPool manages websocket connections and allow sending message
type ConnectionPool struct {
	connections       []*chatsocket.SocketConnection
	connectionsByUser map[string][]*chatsocket.SocketConnection
	connectionByID    map[string]*chatsocket.SocketConnection
	// is used for "write pump"
	sendChannel map[string]chan ([]byte)
}

// NewConnectionPool create new connection pool, ready to use
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections:       make([]*chatsocket.SocketConnection, 0),
		connectionsByUser: make(map[string][]*chatsocket.SocketConnection),
		connectionByID:    make(map[string]*chatsocket.SocketConnection),
		sendChannel:       make(map[string]chan []byte),
	}
}

var _ repository.SocketConnectionRepository = (*ConnectionPool)(nil)
var _ repository.SendMessageRepository = (*ConnectionPool)(nil)

// GetConnectionByUser returns connection ID of all connection of a user
func (pool *ConnectionPool) GetConnectionByUser(userID string) ([]string, error) {
	conns := pool.connectionsByUser[userID]
	result := make([]string, len(conns))
	for i, conn := range conns {
		result[i] = conn.ConnID
	}
	return result, nil
}

// AddConnection resgiter new connection
func (pool *ConnectionPool) AddConnection(conn *chatsocket.SocketConnection) (string, error) {
	conn.ConnID = bson.NewObjectId().Hex()
	// random until it unique
	for {
		if _, exist := pool.connectionByID[conn.ConnID]; exist {
			conn.ConnID = bson.NewObjectId().Hex()
		} else {
			break
		}
	}

	pool.connections = append(pool.connections, conn)
	pool.connectionsByUser[conn.UserID] = append(pool.connectionsByUser[conn.UserID], conn)
	pool.connectionByID[conn.ConnID] = conn
	pool.sendChannel[conn.ConnID] = make(chan []byte, 10)
	go writePump(conn.Conn, pool.sendChannel[conn.ConnID])
	return conn.ConnID, nil
}

// RemoveConnection remove connection with specified ID from all maps
func (pool *ConnectionPool) RemoveConnection(connID string) error {
	var hasRemoved bool
	var removedConn *chatsocket.SocketConnection
	pool.connections, removedConn, hasRemoved = removeConn(connID, pool.connections)
	if !hasRemoved {
		return errors.New("Not Found")
	}
	fmt.Println(pool.connections, removedConn, hasRemoved)
	pool.connectionsByUser[removedConn.UserID], _, _ = removeConn(connID, pool.connectionsByUser[removedConn.UserID])
	delete(pool.connectionByID, connID)
	close(pool.sendChannel[connID])
	delete(pool.sendChannel, connID)
	return nil
}

// SendMessage send message to specifed socket, if it's []byte then call write message, otherwise call writeJSON
func (pool *ConnectionPool) SendMessage(connID string, data interface{}) error {
	_, exist := pool.connectionByID[connID]
	if !exist {
		return errors.New("Connection with that ID not found")
	}
	var messageBytes []byte
	var err error
	switch data.(type) {
	case []byte:
		messageBytes = data.([]byte)
	default:
		messageBytes, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	pool.sendChannel[connID] <- messageBytes
	return nil
}

func removeConn(connID string, connArr []*chatsocket.SocketConnection) ([]*chatsocket.SocketConnection, *chatsocket.SocketConnection, bool) {
	n := len(connArr)
	found := false
	for i := 0; i < n; i++ {
		connArr[i], connArr[n-1] = connArr[n-1], connArr[i]
		found = true
		break
	}
	if found {
		res := connArr[n-1]
		connArr = connArr[:n-1]
		return connArr, res, true
	}
	return connArr, &chatsocket.SocketConnection{}, false
}

func writePump(conn *websocket.Conn, sendChan <-chan []byte) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case message, ok := <-sendChan:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(sendChan)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-sendChan)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
