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
	RoomName         string          `json:"roomName" bson:"roomName,omitempty"`
	OrgID            bson.ObjectId   `json:"orgId" bson:"orgId,omitempty"`
	CreatedTimeStamp time.Time       `json:"-" bson:"createdAt,omitempty"`
	RoomType         string          `json:"roomType" bson:"roomType,omitempty"`
	ListUser         []bson.ObjectId `json:"users" bson:"users,omitempty"`
	ListProxy        []bson.ObjectId `json:"proxies" bson:"proxies,omitempty"`
	ListAdmin        []bson.ObjectId `json:"admins" bson:"admins,omitempty"`
}

// RoomUpdateMongo has same fields as room, but has types of interface{}.
// It's used instead of raw bson.M in update operations to ensure that when field name change in room model
// is always reflected
type RoomUpdateMongo struct {
	RoomID           interface{} `bson:"_id,omitempty"`
	RoomName         interface{} `bson:"roomName,omitempty"`
	OrgID            interface{} `bson:"orgId,omitempty"`
	CreatedTimeStamp interface{} `bson:"createdAt,omitempty"`
	RoomType         interface{} `bson:"roomType,omitempty"`
	ListUser         interface{} `bson:"users,omitempty"`
	ListProxy        interface{} `bson:"proxies,omitempty"`
	ListAdmin        interface{} `json:"admins" bson:"admins,omitempty"`
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
