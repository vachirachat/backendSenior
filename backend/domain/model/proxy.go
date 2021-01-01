package model

import "github.com/globalsign/mgo/bson"

// Proxy represent proxy that connect to controller
type Proxy struct {
	ProxyID bson.ObjectId   `json:"proxyId" bson:"_id,omitempty"`
	IP      string          `json:"ip" bson:"ip,omitempty"`
	Port    int             `json:"port" bson:"port,omitempty"`
	Secret  string          `json:"-" bson:"secret,omitempty"`
	Name    string          `json:"name" bson:"name,omitempty"`
	Rooms   []bson.ObjectId `json:"rooms" bson:"rooms,omitempty"`
}

// ProxyInsert is used for inserting where empty fields are
// not omitted so that we can insert empty array to the database
type ProxyInsert struct {
	ProxyID bson.ObjectId   `json:"proxyId" bson:"_id,omitempty"`
	IP      string          `json:"ip" bson:"ip"`
	Port    int             `json:"port" bson:"port"`
	Secret  string          `json:"-" bson:"secret"`
	Name    string          `json:"name" bson:"name"`
	Rooms   []bson.ObjectId `json:"rooms" bson:"rooms"`
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
