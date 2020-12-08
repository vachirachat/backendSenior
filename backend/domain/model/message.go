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

type Message struct {
	MessageID bson.ObjectId `json:"messageId" bson:"_id,omitempty"`
	TimeStamp time.Time     `json:"timestamp" bson:"timestamp"`
	RoomID    bson.ObjectId `json:"roomId" bson:"roomId"`
	UserID    bson.ObjectId `json:"userId" bson:"userId"`
	Data      string        `json:"data" bson:"data"`
	Type      string        `json:"type" bson:"type"`
}

// TimeRange is used for filtering message by time
type TimeRange struct {
	From time.Time
	To   time.Time
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
