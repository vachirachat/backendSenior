package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// FCMToken store user's Firebase Cloud Messaging Token
// FCM token is used as identity of user device to send message to
type FCMToken struct {
	Token       string        `json:"token" bson:"_id,omitempty"`
	UserID      bson.ObjectId `json:"userId" bson:"userId,omitempty"`
	DeviceName  string        `json:"deviceName" bson:"deviceName,omitempty"`
	LastUpdated time.Time     `json:"lastUpdated" bson:"lastUpdated,omitempty"`
}
