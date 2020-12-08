package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type Conversation struct {
	TimeStamp time.Time     `json:"timestamp" bson:"timestamp"`
	sessionId bson.ObjectId `json:"sessionId" bson:"sessionId"`
	typeCon   string        `json:"typeCon" bson:"typeCon"`
}