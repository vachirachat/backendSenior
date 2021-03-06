package model

import (
	"github.com/globalsign/mgo/bson"
)

type OrganizeInfo struct {
	Orgs []Organize `json:"orgs"`
}

type Organize struct {
	OrganizeID bson.ObjectId   `json:"orgId" bson:"_id,omitempty"`
	Name       string          `json:"name" bson:"name,omitempty"`
	Members    []bson.ObjectId `json:"members" bson:"members,omitempty"`
	Admins     []bson.ObjectId `json:"admins" bson:"admins,omitempty"`
	Rooms      []bson.ObjectId `json:"rooms" bson:"rooms,omitempty"`
	Proxies    []bson.ObjectId `json:"proxies" bson:"proxies,omitempty"`
}

// OrganizationInsert is used for inserting where empty fields are
// not omitted so that we can insert empty array to the database
type OrganizationInsert struct {
	OrganizeID bson.ObjectId   `json:"orgId" bson:"_id,omitempty"`
	Name       string          `json:"name" bson:"name"`
	Members    []bson.ObjectId `json:"members" bson:"members"`
	Admins     []bson.ObjectId `json:"admins" bson:"admins"`
	Rooms      []bson.ObjectId `json:"rooms" bson:"rooms"`
	Proxies    []bson.ObjectId `json:"proxies" bson:"proxies"`
}

// OrganizationT has same fields as organization, but has types of interface{}.
// It's used instead of raw bson.M in update operations to ensure that when field name change in organization model
// is always reflected
type OrganizationT struct {
	OrganizeID interface{} `bson:"_id,omitempty"`
	Name       interface{} `bson:"name,omitempty"`
	Members    interface{} `bson:"members,omitempty"`
	Admins     interface{} `bson:"admins,omitempty"`
	Rooms      interface{} `bson:"rooms,omitempty"`
	Proxies    interface{} `bson:"proxies,omitempty"`
}
