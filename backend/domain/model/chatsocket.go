package model

import "github.com/gorilla/websocket"

// this defines model related to socket

type SocketConnection struct {
	Conn   *websocket.Conn
	ConnID string
	// For query purpose
	UserID string
}
