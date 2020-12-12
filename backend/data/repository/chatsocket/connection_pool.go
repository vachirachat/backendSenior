package chatsocket

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"errors"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/websocket"
)

// ConnectionPool manages websocket connections and allow sending message
type ConnectionPool struct {
	connections       []model.SocketConnection
	connectionsByUser map[string][]model.SocketConnection
	connectionByID    map[string]model.SocketConnection
}

// NewConnectionPool create new connection pool, ready to use
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections:       make([]model.SocketConnection, 7),
		connectionsByUser: make(map[string][]model.SocketConnection),
		connectionByID:    make(map[string]model.SocketConnection),
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
func (pool *ConnectionPool) AddConnection(conn model.SocketConnection) (string, error) {
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
	pool.connectionByID[conn.UserID] = conn
	return conn.ConnID, nil
}

// RemoveConnection remove connection with specified ID from all maps
func (pool *ConnectionPool) RemoveConnection(connID string) error {
	var hasRemoved bool
	var removedConn model.SocketConnection
	pool.connections, removedConn, hasRemoved = removeConn(connID, pool.connections)
	if !hasRemoved {
		return errors.New("Not Found")
	}
	pool.connectionsByUser[removedConn.UserID], _, _ = removeConn(connID, pool.connectionsByUser[removedConn.UserID])
	delete(pool.connectionByID, connID)
	return nil
}

// SendMessage send message to specifed socket, if it's []byte then call write message, otherwise call writeJSON
func (pool *ConnectionPool) SendMessage(connID string, data interface{}) error {
	connModel, exist := pool.connectionByID[connID]
	if !exist {
		return errors.New("Connection with that ID not found")
	}
	conn := connModel.Conn
	switch data.(type) {
	case []byte:
		conn.WriteMessage(websocket.BinaryMessage, data.([]byte))
	default:
		conn.WriteJSON(data)
	}
	return nil
}

func removeConn(connID string, connArr []model.SocketConnection) ([]model.SocketConnection, model.SocketConnection, bool) {
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
	return connArr, model.SocketConnection{}, false
}
