package model

import (
	"errors"
	"time"

	"github.com/globalsign/mgo/bson"
)

// MessagesResponse is struct for return array of messages
type MessagesResponse struct {
	Messages []Message `json:"messages"`
}

// type Message struct {
// 	MessageID bson.ObjectId `json:"messageId" bson:"_id,omitempty"`
// 	TimeStamp time.Time     `json:"timestamp" bson:"timestamp"`
// 	RoomID    bson.ObjectId `json:"roomId" bson:"roomId"`
// 	UserID    bson.ObjectId `json:"userId" bson:"userId"`
// 	Data      string        `json:"data" bson:"data"`
// 	Type      string        `json:"type" bson:"type"`
// }

type Message struct {
	MessageID string    `json:"messageId" bson:"_id,omitempty"`
	TimeStamp time.Time `json:"timestamp" bson:"timestamp"`
	RoomID    string    `json:"roomId" bson:"roomId"`
	UserID    string    `json:"userId" bson:"userId"`
	Data      string    `json:"data" bson:"data"`
	Type      string    `json:"type" bson:"type"`
}

// TimeRange is used for filtering message by time
type TimeRange struct {
	From time.Time `json:"from" bson:"from"`
	To   time.Time `json:"to" bson:"to"`
}

// Fill replace From with epoch zero and fill To with currentTime
func (rng *TimeRange) Fill() {
	if rng.From.IsZero() {
		rng.From = time.Unix(0, 0)
	}
	if rng.To.IsZero() {
		rng.To = time.Now()
	}
}

// Filled return copy of time, filled
func (rng TimeRange) Filled() TimeRange {
	if rng.From.IsZero() {
		rng.From = time.Unix(0, 0)
	}
	if rng.To.IsZero() {
		rng.To = time.Now()
	}
	return rng
}

// NewDefaultTimeRange  return time range from epoch zero to now
func NewDefaultTimeRange() TimeRange {
	rng := TimeRange{}
	rng.Fill()
	return rng
}

// Validate return whether time range is valid
func (rng *TimeRange) Validate() error {

	if rng.From.IsZero() || rng.To.IsZero() {
		return nil
	}
	if rng.To.Before(rng.From) {
		return errors.New("From must be less than To")
	}
	return nil
}

// Re-Assign byte string(From mondo bson.ObjectID) to String
func (msg *Message) MessageStringIDToMongoID() Message {
	msg.MessageID = bson.ObjectId(msg.MessageID).Hex()
	msg.RoomID = bson.ObjectId(msg.RoomID).Hex()
	msg.UserID = bson.ObjectId(msg.UserID).Hex()
	return *msg
}

func MessageListToMongoID(messages []Message) []Message {
	for i := range messages {
		messages[i] = messages[i].MessageStringIDToMongoID()
	}
	return messages
}
