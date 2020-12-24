package model_proxy

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// KeyRecord represent key stored in database
type KeyRecord struct {
	Key       []byte    `json:"key" bson:"key"`
	ValidFrom time.Time `json:"from" bson:"from"`
	ValidTo   time.Time `json:"to" bson:"to"`
}

type RoomKeys struct {
	RoomID     bson.ObjectId `json:"room" bson:"_id"`
	KeyRecodes []KeyRecord   `json:"keys" bson:"keys"`
}
