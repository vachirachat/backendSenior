package model

import (
	"github.com/globalsign/mgo/bson"
)

type OrganizeInfo struct {
	Orgs []Organize `json:"orgs"`
}

type Organize struct {
	OrganizeID bson.ObjectId   `json:"orgId" bson:"_id"`
	Name       string          `json:"name" bson:"name"`
	Members    []bson.ObjectId `json:"members" bson:"members"`
	Admins     []bson.ObjectId `json:"admins" bson:"admins"`
}
