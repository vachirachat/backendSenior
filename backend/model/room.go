package model

import (
	"time"
)

type RoomInfo struct {
	Room []Room `json:"rooms"`
}

type Room struct {
	RoomID           string    `json:"roomId" bson:"roomId"`
	RoomName         string    `json:"roomName" bson:"roomName"`
	CreatedTimeStamp time.Time `json:"createdTimestamp" bson:"createdTimestamp"`
	RoomType         string    `json:"roomType" bson:"roomType"`
	ListUser         []string  `json:"listUser" bson:"listUser"`
}
