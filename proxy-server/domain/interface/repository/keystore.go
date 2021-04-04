package repository

import (
	model_proxy "proxySenior/domain/model"
	"time"
)

type Keystore interface {
	GetKeyForMessage(roomID string, timestamp time.Time) (key model_proxy.KeyRecord, err error)
	GetKeyByRoom(roomID string) (keys []model_proxy.KeyRecord, err error)
	AddNewKey(model_proxy.KeyRecord)
	RequestKeyForMessage(roomID string, timestamp time.Time) (key model_proxy.KeyRecord, err error)
}
