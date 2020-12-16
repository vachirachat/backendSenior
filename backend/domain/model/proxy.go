package model

import "github.com/globalsign/mgo/bson"

// Proxy represent proxy that connect to controller
type Proxy struct {
	ProxyID bson.ObjectId   `json:"proxyId" bson:"_id,omitempty"`
	Name    string          `json:"name" bson:"name,omitempty"`
	Rooms   []bson.ObjectId `json:"rooms" bson:"rooms,omitempty"`
}
