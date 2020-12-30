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

// ProxyUpdateMongo has same fields as proxy, but has types of interface{}.
// It's used instead of raw bson.M in update operations to ensure that when field name change in proxy model
// is always reflected
type ProxyUpdateMongo struct {
	ProxyID interface{} `bson:"_id,omitempty"`
	IP      interface{} `bson:"ip,omitempty"`
	Port    interface{} `bson:"port,omitempty"`
	Secret  interface{} `bson:"secret,omitempty"`
	Name    interface{} `bson:"name,omitempty"`
	Rooms   interface{} `bson:"rooms,omitempty"`
}
