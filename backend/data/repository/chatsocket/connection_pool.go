package chatsocket

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model/chatsocket"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
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
	connections       []*chatsocket.Connection
	connectionsByUser map[string][]*chatsocket.Connection
	connectionByID    map[string]*chatsocket.Connection
	// is used for "write pump"
	sendChannel map[string]chan ([]byte)
	lock        sync.RWMutex
}

// NewConnectionPool create new connection pool, ready to use
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections:       make([]*chatsocket.Connection, 0),
		connectionsByUser: make(map[string][]*chatsocket.Connection),
		connectionByID:    make(map[string]*chatsocket.Connection),
		sendChannel:       make(map[string]chan []byte),
		lock:              sync.RWMutex{},
	}
}

var _ repository.SocketConnectionRepository = (*ConnectionPool)(nil)
var _ repository.SendMessageRepository = (*ConnectionPool)(nil)

// GetConnectionByUser returns connection ID of all connection of a user
func (pool *ConnectionPool) GetConnectionByUser(userID string) ([]string, error) {
	pool.lock.RLock()
	conns := pool.connectionsByUser[userID]
	pool.lock.RUnlock()

	result := make([]string, len(conns))
	for i, conn := range conns {
		result[i] = conn.ConnID
	}
	return result, nil
}

// AddConnection resgiter new connection
func (pool *ConnectionPool) AddConnection(conn *chatsocket.Connection) (string, error) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

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
	var removedConn *chatsocket.Connection

	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.connections, removedConn, hasRemoved = removeConn(connID, pool.connections)
	if !hasRemoved {
		return errors.New("Not Found")
	}
	pool.connectionsByUser[removedConn.UserID], _, _ = removeConn(connID, pool.connectionsByUser[removedConn.UserID])
	delete(pool.connectionByID, connID)
	close(pool.sendChannel[connID])
	delete(pool.sendChannel, connID)

	return nil
}

// SendMessage send message to specifed socket, if it's []byte then call write message, otherwise call writeJSON
func (pool *ConnectionPool) SendMessage(connID string, data interface{}) error {
	pool.lock.RLock()
	_, exist := pool.connectionByID[connID]
	sendChan := pool.sendChannel[connID] // might be nil
	pool.lock.RUnlock()

	if !exist {
		fmt.Println("[send message] conn", connID, "not found")
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
	sendChan <- messageBytes
	return nil
}

func removeConn(connID string, connArr []*chatsocket.Connection) ([]*chatsocket.Connection, *chatsocket.Connection, bool) {
	n := len(connArr)
	found := false
	for i := 0; i < n; i++ {
		if connArr[i].ConnID == connID {
			connArr[i], connArr[n-1] = connArr[n-1], connArr[i]
			found = true
			break
		}
	}
	if found {
		res := connArr[n-1]
		connArr = connArr[:n-1]
		return connArr, res, true
	}
	return connArr, &chatsocket.Connection{}, false
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
