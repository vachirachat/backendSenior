package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type MessageInfo struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	MessageID bson.ObjectId `json:"messageId" bson:"_id,omitempty"`
	TimeStamp time.Time     `json:"timestamp" bson:"timestamp"`
	RoomID    string        `json:"roomID" bson:"roomID"`
	UserID    string        `json:"userID" bson:"userID"`
	Data      string        `json:"data" bson:"data"`
	Type      string        `json:"type" bson:"type"`
}
