package model

import "github.com/globalsign/mgo/bson"

// KeyVersion used to control version in key exchange
type KeyVersion struct {
	RoomID   bson.ObjectId `json:"roomId" bson:"roomId" `
	ProxyID  bson.ObjectId `json:"proxyId" bson:"proxyId"`
	Priority int           `json:"priority" bson:"priority"`
	Version  int           `json:"version" bson:"version"`
}

// KeyVersionFilter used for filtering key
type KeyVersionFilter struct {
	RoomID   interface{} `bson:"roomId,omitempty" `
	ProxyID  interface{} `bson:"proxyId,omitempty"`
	Priority int         `bson:"priority,omitempty"`
	Version  interface{} `bson:"version,omitempty"`
}
