package dto

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

// FindMessageDto is for find message (currently) by time
type FindMessageDto struct {
	RoomID bson.ObjectId `validate:"required"`
	From   time.Time
	To     time.Time
}
