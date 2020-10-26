package model

import (
	"time"
)

type Room struct {
	RoomID           string    `json:"roomId" bson:"roomId"`
	RoomName         string    `json:"roomName" bson:"roomName"`
	CreatedTimeStamp time.Time `json:"createdTimestamp" bson:"createdTimestamp"`
	UserID           string    `json:"userId" bson:"userId"`
	SocketID         string    `json:"socketid" bson:"socketid"`
	AdmitID          []string  `json:"adminid" bson:"adminid"`
	MessageQueue     []Message `json:"MessageQueue" bson:"MessageQueue"`
}
