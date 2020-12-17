package repository

import (
	"proxySenior/domain/model"
	"time"
)

type Keystore interface {
	GetKeyForMessage(roomID string, timestamp time.Time) (key model.KeyRecord, err error)
	GetKeyByRoom(roomID string) (keys []model.KeyRecord, err error)
	AddNewKey(model.KeyRecord)

	RequestKeyForMessage(roomID string, timestamp time.Time) (key model.KeyRecord, err error)
}
