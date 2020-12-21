package model

import (
	"github.com/globalsign/mgo/bson"
)

type OrganizeInfo struct {
	Organize []Organize
}

type Organize struct {
	OrganizeID bson.ObjectId   `json:"organizeId" bson:"_id"`
	Name       string          `json:"name" bson:"name"`
	ListMember []bson.ObjectId `json:"listMember" bson:"listMember"`
	ListAdmin  []bson.ObjectId `json:"listAdmin" bson:"listAdmin"`
}
