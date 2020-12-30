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
}

// OrganizationUpdateMongo has same fields as organization, but has types of interface{}.
// It's used instead of raw bson.M in update operations to ensure that when field name change in organization model
// is always reflected
type OrganizationUpdateMongo struct {
	OrganizeID interface{} `bson:"_id,omitempty"`
	Name       interface{} `bson:"name,omitempty"`
	Members    interface{} `bson:"members,omitempty"`
	Admins     interface{} `bson:"admins,omitempty"`
}
