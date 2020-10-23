package model

import (
	"time"
)

type Message struct {
	TimeStamp time.Time `json:"timestamp" bson:"timestamp"`
	RoomID    string    `json:"roomID" bson:"roomID"`
	UserID    string    `json:"userID" bson:"userID"`
	Data      string    `json:"data" bson:"data"`
	Type      string    `json:"type" bson:"type"`
}
