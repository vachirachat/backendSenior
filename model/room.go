package model

import (
	"time"
)

type Room struct {
	RoomID           string    `json:"room_id" bson:"room_id"`
	RoomName         string    `json:"room_name" bson:"room_name"`
	CreatedTimeStamp time.Time `json:"created_timestamp" bson:"created_timestamp"`
	UserID           string    `json:"user_id" bson:"user_id"`
	SocketID         string    `json:"socketid" bson:"socketid"`
	AdmitID          []string  `json:"adminid" bson:"adminid"`
	MessageQueue     []Message `json:"MessageQueue" bson:"MessageQueue"`
}
