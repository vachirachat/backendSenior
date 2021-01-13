package repository

import (
	"backendSenior/domain/model/chatsocket/key_exchange"
	model_proxy "proxySenior/domain/model"
)

// Keystore represent local keystore
type Keystore interface {
	// Find will find key according to filter
	Find(filter interface{}) ([]model_proxy.KeyRecord, error)

	// FindByRoom is shortcut for finding keys by room
	FindByRoom(roomID string) ([]model_proxy.KeyRecord, error)

	// AddNewKey should add a key to room, while invalidate the last key (if exists)
	AddNewKey(roomID string, key []byte) error

	// Note that key can't be deleted
}

// RemoteKeyStore represent remote key store
type RemoteKeyStore interface {
	GetByRoom(roomID string, details key_exchange.KeyExchangeRequest) (key_exchange.KeyExchangeResponse, error)
}
