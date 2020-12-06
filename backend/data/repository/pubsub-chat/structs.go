package repository

import (
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/websocket"
)

// UserStruct is used for sending users with socket id
type UserStruct struct {
	Username string        `json:"username"`
	UserID   bson.ObjectId `json:"userID"`
}

// SocketEventStruct struct of socket events
type SocketEventStruct struct {
	EventName    string      `json:"eventName"`
	EventPayload interface{} `json:"eventPayload"`
}

type messagePayload struct {
	UserId    bson.ObjectId `json:"userId"`
	RoomId    bson.ObjectId `json:"roomId"`
	Message   string        `json:"message"`
	Timestamp time.Time     `json:"timestamp"`
}

type SocketMessageEventStruct struct {
	EventName    string         `json:"eventName"`
	EventPayload messagePayload `json:"eventPayload"`
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub                 *Hub
	webSocketConnection *websocket.Conn
	send                chan SocketEventStruct
	username            string
	userID              bson.ObjectId

	Room []bson.ObjectId //add
}

// JoinDisconnectPayload will have struct for payload of join disconnect
type JoinDisconnectPayload struct {
	Users  []UserStruct  `json:"users"`
	UserID bson.ObjectId `json:"userID"`
}

type RoomPayload struct {
	RoomId bson.ObjectId
	UserID bson.ObjectId `json:"userID"`
}
