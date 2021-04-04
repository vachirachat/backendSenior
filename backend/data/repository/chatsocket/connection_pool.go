package chatsocket

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model/chatsocket"
	"errors"
	"fmt"
	"sync"

	"github.com/globalsign/mgo/bson"
)

// ConnectionPool manages websocket connections and allow sending message
type ConnectionPool struct {
	connections       []*chatsocket.Connection
	connectionsByUser map[string][]*chatsocket.Connection
	connectionByID    map[string]*chatsocket.Connection
	lock              sync.RWMutex
}

// NewConnectionPool create new connection pool, ready to use
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections:       make([]*chatsocket.Connection, 0),
		connectionsByUser: make(map[string][]*chatsocket.Connection),
		connectionByID:    make(map[string]*chatsocket.Connection),
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

// AddConnection register new connection
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
		return errors.New("not Found")
	}
	pool.connectionsByUser[removedConn.UserID], _, _ = removeConn(connID, pool.connectionsByUser[removedConn.UserID])
	delete(pool.connectionByID, connID)

	return nil
}

// SendMessage send message to specified socket, if it's []byte then call write message, otherwise call writeJSON
func (pool *ConnectionPool) SendMessage(connID string, data interface{}) error {
	pool.lock.RLock()
	conn, exist := pool.connectionByID[connID]
	pool.lock.RUnlock()

	if !exist {
		fmt.Println("[send message] conn", connID, "not found")
		return errors.New("connection with that ID not found")
	}
	switch v := data.(type) {
	case []byte:
		return conn.Conn.Send(v)
	default:
		return conn.Conn.SendJSON(v)
	}
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

func (pool *ConnectionPool) DebugNumOfConns() map[string]interface{} {
	res := make(map[string]interface{})
	res["total"] = len(pool.connections)
	by_user := make(map[string]int)
	for u, conns := range pool.connectionsByUser {
		by_user[u] = len(conns)
	}
	res["by_user"] = by_user
	return res
}
