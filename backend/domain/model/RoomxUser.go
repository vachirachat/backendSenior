package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type RoomxUser struct {
	TimeStamp time.Time       `json:"timestamp" bson:"timestamp"`
	RoomID    bson.ObjectId   `json:"roomID" bson:"roomID"`
	UserID    []bson.ObjectId `json:"userID" bson:"userID"`
}