package service

import (
	"crypto/rand"
	"proxySenior/domain/interface/repository"
	model_proxy "proxySenior/domain/model"
)

// KeyService is service for managing key
type KeyService struct {
	rep repository.Keystore // local
}

// NewKeyService create new key service
func NewKeyService(rep repository.Keystore) *KeyService {
	return &KeyService{
		rep: rep,
	}
}

// ---- local

// GetKeyForRoom is used for getting keys for room locally
func (s *KeyService) GetKeyForRoom(roomID string) ([]model_proxy.KeyRecord, error) {
	keys, err := s.rep.Find(roomID)
	return keys, err
}

// NewKeyForRoom generate new key for room, invalidating old one
func (s *KeyService) NewKeyForRoom(roomID string) error {
	key, err := randomBytes(32)
	if err != nil {
		return err
	}

	err = s.rep.AddNewKey(roomID, key)

	return err
}

// IsLocal
func (s *KeyService) IsLocal(roomID string) (bool, error) {
	panic("todo")
}

// TODO: use other way to generate
// generate key, size should be 32
func randomBytes(size int) ([]byte, error) {
	key := make([]byte, size)
	n, err := rand.Read(key)
	if err != nil || n != size {
		return nil, err
	}
	return key, err
}
