package model

import (
	"time"
)

type Conversation struct {
	TimeStamp time.Time `json:"timestamp" bson:"timestamp"`
	sessionId string    `json:"sessionId" bson:"sessionId"`
	typeCon   string    `json:"typeCon" bson:"typeCon"`
}

type ConversationDev struct {
	TimeStamp time.Time `json:"timestamp" bson:"timestamp"`
	sessionId string    `json:"sessionId" bson:"sessionId"`
	typeCon   string    `json:"typeCon" bson:"typeCon"`
	Email     string    `json:"typeCon" bson:"typeCon"`
}
