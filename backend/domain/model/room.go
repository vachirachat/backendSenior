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
	CreatedTimeStamp time.Time       `json:"-" bson:"createdAt"`
	RoomType         string          `json:"roomType" bson:"roomType"`
	ListUser         []bson.ObjectId `json:"users" bson:"users"`
	ListProxy        []bson.ObjectId `json:"proxies" bson:"proxies"`
}

// Map return Room struct as Map
func (room *Room) Map() map[string]interface{} {
	return map[string]interface{}{
		"_id":       room.RoomID,
		"roomName":  room.RoomName,
		"createdAt": room.CreatedTimeStamp,
		"roomType":  room.RoomType,
		"users":     room.ListUser,
		"proxies":   room.ListProxy,
	}
}
