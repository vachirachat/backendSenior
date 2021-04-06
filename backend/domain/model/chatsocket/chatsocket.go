package chatsocket

import (
	"backendSenior/domain/model/chatsocket/message_types"
	"common/ws"
	"encoding/json"
)

// this defines model related to socket

type Connection struct {
	Conn   *ws.Connection
	ConnID string
	// For query purpose
	UserID string // or proxy ID
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

// InvalidateRoomMasterMessage create message for invalidating room master
func InvalidateRoomMasterMessage(roomID string) Message {
	return Message{
		Type:    message_types.InvalidateMaster,
		Payload: roomID,
	}
}

// InvalidateRoomKeyMessage crate message for invalidating room key
func InvalidateRoomKeyMessage(roomID string) Message {
	return Message{
		Type:    message_types.InvalidateKey,
		Payload: roomID,
	}
}
