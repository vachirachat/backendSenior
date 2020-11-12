package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/websocket"
)

type Conversation struct {
	TimeStamp time.Time     `json:"timestamp" bson:"timestamp"`
	sessionId bson.ObjectId `json:"sessionId" bson:"sessionId"`
	typeCon   string        `json:"typeCon" bson:"typeCon"`
}

type Hub struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
}

// UserStruct is used for sending users with socket id
type UserStruct struct {
	Username string `json:"username"`
	UserID   string `json:"userID"`
}

// SocketEventStruct struct of socket events
type SocketEventStruct struct {
	EventName    string      `json:"eventName"`
	EventPayload interface{} `json:"eventPayload"`
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub                 *Hub
	WebSocketConnection *websocket.Conn
	Send                chan SocketEventStruct
	Username            string
	UserID              string
}

// JoinDisconnectPayload will have struct for payload of join disconnect
type JoinDisconnectPayload struct {
	Users  []UserStruct `json:"users"`
	UserID string       `json:"userID"`
}
