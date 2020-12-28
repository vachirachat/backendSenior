package chatsocket

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

// this defines model related to socket

type Connection struct {
	Conn   *websocket.Conn
	ConnID string
	// For query purpose
	UserID string
}

// RawMessage is SocketMessage type designed to be parsed by JSON
type RawMessage struct {
	Type    string          `json:"type"`    // category of message
	Payload json.RawMessage `json:"payload"` // skip parsing data until type is known
}

// Message defines generic type of message over websocket
type Message struct {
	Type    string      `json:"type"`    // category of message
	Payload interface{} `json:"payload"` // skip parsing data until type is known
}
