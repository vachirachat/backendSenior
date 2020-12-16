package model

import "time"

// KeyRecord represent key stored in database
type KeyRecord struct {
	Key       []byte    `json:"key" bson:"key"`
	ValidFrom time.Time `json:"from" bson:"from"`
	ValidTo   time.Time `json:"from" bson:"to"`
}
