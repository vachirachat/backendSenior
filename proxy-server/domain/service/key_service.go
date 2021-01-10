package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"proxySenior/domain/interface/repository"
	model_proxy "proxySenior/domain/model"

	"github.com/globalsign/mgo/bson"
)

// KeyService is service for managing key
type KeyService struct {
	rep      repository.Keystore // local
	proxy    repository.ProxyMasterAPI
	clientID string // clientID of this proxy
}

// NewKeyService create new key service
func NewKeyService(rep repository.Keystore, proxy repository.ProxyMasterAPI, clientID string) *KeyService {
	return &KeyService{
		rep:      rep,
		proxy:    proxy,
		clientID: clientID,
	}
}

// ---- local

// GetKeyForRoom is used for getting keys for room locally
func (s *KeyService) GetKeyForRoom(roomID string) ([]model_proxy.KeyRecord, error) {
	if ok, err := s.IsLocal(roomID); err != nil {
		return nil, fmt.Errorf("error checking locality of key: %v", err)
	} else if !ok {
		// TODO
		return nil, errors.New("get key for remote proxy not supported yet")
	}

	keys, err := s.rep.Find(model_proxy.KeyRecordUpdate{
		RoomID: bson.ObjectIdHex(roomID),
	})
	if err == nil && keys == nil {
		keys = []model_proxy.KeyRecord{}
	}

	return keys, err
}

// NewKeyForRoom generate new key for room, invalidating old one
func (s *KeyService) NewKeyForRoom(roomID string) error {
	if ok, err := s.IsLocal(roomID); err != nil {
		return fmt.Errorf("error checking locality of key: %v", err)
	} else if !ok {
		return errors.New("can't generate key for remote proxy")
	}

	key, err := randomBytes(32)
	if err != nil {
		return err
	}

	err = s.rep.AddNewKey(roomID, key)

	return err
}

// IsLocal return whether key from `roomID` can be fetched locally (by key store)
func (s *KeyService) IsLocal(roomID string) (bool, error) {
	proxy, err := s.proxy.GetRoomMasterProxy(roomID)
	if err != nil {
		return false, err
	}
	return proxy.ProxyID.Hex() == s.clientID, nil
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
