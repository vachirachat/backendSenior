package repository

import (
	model_proxy "proxySenior/domain/model"
)

type Keystore interface {
	// Find will find key according to filter
	Find(filter interface{}) ([]model_proxy.KeyRecord, error)

	// FindByRoom is shortcut for finding keys by room
	FindByRoom(roomID string) ([]model_proxy.KeyRecord, error)

	// AddNewKey should add a key to room, while invalidate the last key (if exists)
	AddNewKey(roomID string, key []byte) error

	// Note that key can't be deleted
}
