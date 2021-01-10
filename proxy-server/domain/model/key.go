package model_proxy

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// KeyRecord represent single key for a room for specific time range
type KeyRecord struct {
	ID        bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Key       []byte        `json:"key" bson:"key"`
	RoomID    bson.ObjectId `json:"room" bson:"_id"`
	ValidFrom time.Time     `json:"from" bson:"from"`
	ValidTo   time.Time     `json:"to" bson:"to"`
}

// KeyRecordUpdate is used for updating
type KeyRecordUpdate struct {
	Key       interface{} `json:"key" bson:"key,omitempty"`
	RoomID    interface{} `json:"room" bson:"_id,omitempty"`
	ValidFrom interface{} `json:"from" bson:"from,omitempty"`
	ValidTo   interface{} `json:"to" bson:"to,omitempty"`
}
