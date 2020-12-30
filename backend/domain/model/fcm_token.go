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

// FcmTokenUpdateMongo has same fields as fcmToken, but has types of interface{}.
// It's used instead of raw bson.M in update operations to ensure that when field name change in fcmToken model
// is always reflected
type FcmTokenUpdateMongo struct {
	Token       interface{} `bson:"_id,omitempty"`
	UserID      interface{} `bson:"userId,omitempty"`
	DeviceName  interface{} `bson:"deviceName,omitempty"`
	LastUpdated interface{} `bson:"lastUpdated,omitempty"`
}
