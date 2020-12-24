package model

import "github.com/globalsign/mgo/bson"

// Proxy represent proxy that connect to controller
type Proxy struct {
	ProxyID bson.ObjectId   `json:"proxyId,omitempty" bson:"_id,omitempty"`
	IP      string          `json:"ip,omitempty" bson:"ip,omitempty"`
	Port    int             `json:"port,omitempty" bson:"port,omitempty"`
	Secret  string          `json:"-" bson:"secret,omitempty"`
	Name    string          `json:"name,omitempty" bson:"name,omitempty"`
	Rooms   []bson.ObjectId `json:"rooms,omitempty" bson:"rooms,omitempty"`
}
