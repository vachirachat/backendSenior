package model_proxy

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
}

// RoomInsert is used for inserting where empty fields are
// not omitted so that we can insert empty array to the database
type RoomInsert struct {
	RoomID           bson.ObjectId   `json:"roomId" bson:"_id,omitempty"`
	RoomName         string          `json:"roomName" bson:"roomName"`
	OrgID            bson.ObjectId   `json:"orgId" bson:"orgId,omitempty"` // orgId can be empty (when creating)
	CreatedTimeStamp time.Time       `json:"-" bson:"createdAt"`
	RoomType         string          `json:"roomType" bson:"roomType"`
	ListUser         []bson.ObjectId `json:"users" bson:"users"`
	ListProxy        []bson.ObjectId `json:"proxies" bson:"proxies"`
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
