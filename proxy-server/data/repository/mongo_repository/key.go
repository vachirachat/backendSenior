package mongo_repository

import (
	"proxySenior/domain/interface/repository"
	model_proxy "proxySenior/domain/model"
	"time"
)

type KeyRepository struct {
	repository.Keystore
}

var _ repository.Keystore = (*KeyRepository)(nil)

func (repo *KeyRepository) GetKeyForMessage(roomID string, timestamp time.Time) (model_proxy.KeyRecord, error) {
	return model_proxy.KeyRecord{
		Key: []byte("1234567890abcdef"),
	}, nil
}
