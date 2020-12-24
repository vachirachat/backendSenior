package repository

import (
	model_proxy "proxySenior/domain/model"
	"time"
)

type Keystore interface {
	GetKeyForMessage(roomID string, timestamp time.Time) (key model_proxy.RoomKeys, err error)
	GetKeyByRoom(roomID string) (keys []model_proxy.KeyRecord, err error)
	AddNewKey(roomID string, keyRecord []model_proxy.KeyRecord) (model_proxy.RoomKeys, error)
	UpdateNewKey(roomID string, keyRecord []model_proxy.KeyRecord, timestamp time.Time) (model_proxy.RoomKeys, error)
	// RequestKeyForMessage(roomID string, timestamp time.Time) (key model_proxy.KeyRecord, err error)
}
