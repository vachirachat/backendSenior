package chatsocket

import (
	"backendSenior/domain/model/chatsocket"
	"errors"
)

// worker is the loop that take commands from channel then execute and send its result
func (pool *ConnectionPool) worker() {
	var hasRemoved bool
	var removedConn *chatsocket.Connection

	for {
		select {
		case read := <-pool.readCmdChan:
			read.result <- pool.connectionsByUser[read.userID]
			read.err <- nil

		case add := <-pool.addCmdChan:
			conn := add.conn

			pool.connections = append(pool.connections, conn)
			pool.connectionsByUser[conn.UserID] = append(pool.connectionsByUser[conn.UserID], conn)
			pool.connectionByID[conn.ConnID] = conn
			pool.sendChannel[conn.ConnID] = make(chan []byte, 10)
			go writePump(conn.Conn, pool.sendChannel[conn.ConnID])

			add.err <- nil

		case del := <-pool.delCmdChan:
			connID := del.connID
			pool.connections, removedConn, hasRemoved = removeConn(connID, pool.connections)
			if !hasRemoved {
				del.err <- errors.New("Not Found")
				continue
			}
			pool.connectionsByUser[removedConn.UserID], _, _ = removeConn(connID, pool.connectionsByUser[removedConn.UserID])
			delete(pool.connectionByID, connID)
			close(pool.sendChannel[connID])
			delete(pool.sendChannel, connID)
			del.err <- nil
		}
	}

}
