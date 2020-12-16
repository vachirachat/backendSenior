package mongo_repository

import (
	"proxySenior/domain/interface/repository"
	"proxySenior/domain/model"
	"time"
)

type KeyRepository struct {
	repository.Keystore
}

var _ repository.Keystore = (*KeyRepository)(nil)

func (repo *KeyRepository) GetKeyForMessage(roomID string, timestamp time.Time) (model.KeyRecord, error) {
	return model.KeyRecord{
		Key: []byte("1234567890abcdef"),
	}, nil
}
