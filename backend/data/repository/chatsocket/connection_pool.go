package chatsocket

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model/chatsocket"
	"encoding/json"
	"errors"
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
	// synchronize map
	readCmdChan chan readCmd
	addCmdChan  chan addCmd
	delCmdChan  chan deleteCmd
}

// NewConnectionPool create new connection pool, ready to use
func NewConnectionPool() *ConnectionPool {
	pool := &ConnectionPool{
		connections:       make([]*chatsocket.Connection, 0),
		connectionsByUser: make(map[string][]*chatsocket.Connection),
		connectionByID:    make(map[string]*chatsocket.Connection),
		sendChannel:       make(map[string]chan []byte),
		readCmdChan:       make(chan readCmd, 10),
		addCmdChan:        make(chan addCmd, 10),
		delCmdChan:        make(chan deleteCmd, 10),
	}

	go pool.worker()

	return pool
}

var _ repository.SocketConnectionRepository = (*ConnectionPool)(nil)
var _ repository.SendMessageRepository = (*ConnectionPool)(nil)

// GetConnectionByUser returns connection ID of all connection of a user
func (pool *ConnectionPool) GetConnectionByUser(userID string) ([]string, error) {
	ret := make(chan []*chatsocket.Connection, 1)
	err := make(chan error, 1)

	pool.readCmdChan <- readCmd{userID, ret, err}

	userConns := <-ret
	result := make([]string, len(userConns))
	for i, conn := range userConns {
		result[i] = conn.ConnID
	}

	return result, <-err
}

// AddConnection resgiter new connection
func (pool *ConnectionPool) AddConnection(conn *chatsocket.Connection) (string, error) {
	conn.ConnID = bson.NewObjectId().Hex()
	// random until it unique
	for {
		if _, exist := pool.connectionByID[conn.ConnID]; exist {
			conn.ConnID = bson.NewObjectId().Hex()
		} else {
			break
		}
	}

	err := make(chan error, 1)
	pool.addCmdChan <- addCmd{conn, err}

	return conn.ConnID, <-err
}

// RemoveConnection remove connection with specified ID from all maps
func (pool *ConnectionPool) RemoveConnection(connID string) error {
	err := make(chan error)
	pool.delCmdChan <- deleteCmd{connID, err}
	return <-err
}

// SendMessage send message to specifed socket, if it's []byte then call write message, otherwise call writeJSON
func (pool *ConnectionPool) SendMessage(connID string, data interface{}) error {
	_, exist := pool.connectionByID[connID]
	if !exist {
		return errors.New("Connection with that ID not found")
	}

	messageBytes, err := toBytes(data)
	if err != nil {
		return err
	}

	pool.sendChannel[connID] <- messageBytes
	return nil
}

func toBytes(data interface{}) ([]byte, error) {
	var messageBytes []byte
	var err error
	switch data.(type) {
	case []byte:
		messageBytes = data.([]byte)
	default:
		messageBytes, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}
	return messageBytes, nil
}

func removeConn(connID string, connArr []*chatsocket.Connection) ([]*chatsocket.Connection, *chatsocket.Connection, bool) {
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
