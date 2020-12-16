package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type RoomInfo struct {
	Room []Room `json:"rooms"`
}

type Room struct {
	RoomID           bson.ObjectId   `json:"roomId" bson:"_id,omitempty"`
	RoomName         string          `json:"roomName" bson:"roomName"`
	CreatedTimeStamp time.Time       `json:"-" bson:"createdTimestamp"`
	RoomType         string          `json:"roomType" bson:"roomType"`
	ListUser         []bson.ObjectId `json:"listUser" bson:"listUser"`
	ListProxy        []bson.ObjectId `json:"listProxy" bson:"listProxy"`
}

// Map return Room struct as Map
func (room *Room) Map() map[string]interface{} {
	return map[string]interface{}{
		"roomId":    room.RoomID,
		"roomName":  room.RoomName,
		"timestamp": room.CreatedTimeStamp,
		"roomType":  room.RoomType,
		"listUser":  room.ListUser,
		"listProxy": room.ListProxy,
	}
}
