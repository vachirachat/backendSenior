package model

import (
	"time"
)

type RoomxUser struct {
	TimeStamp time.Time `json:"timestamp" bson:"timestamp"`
	RoomID    string    `json:"roomID" bson:"roomID"`
	UserID    []string  `json:"userID" bson:"userID"`
}
