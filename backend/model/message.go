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
	RoomID    bson.ObjectId `json:"roomID" bson:"roomID"`
	UserID    bson.ObjectId `json:"userID" bson:"userID"`
	Name      string        `json:"username" bson:"username"`
	Data      string        `json:"data" bson:"data"`
	Type      string        `json:"type" bson:"type"`
}
