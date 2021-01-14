package key_service

// manage local key

import (
	"errors"
	"fmt"
	model_proxy "proxySenior/domain/model"
	"time"

	"github.com/globalsign/mgo/bson"
)

// GetKeyLocal is used for getting keys for room locally if possible
func (s *KeyService) GetKeyLocal(roomID string) ([]model_proxy.KeyRecord, error) {
	ok, err := s.IsLocal(roomID)
	if err != nil {
		return nil, fmt.Errorf("error checking locality of key: %v", err)
	}

	if !ok {
		return nil, errors.New("can't get local key for remote room")
	}

	keys, err := s.local.Find(model_proxy.KeyRecordUpdate{
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

	err = s.local.AddNewKey(roomID, key)

	return err
}

// IsLocal return whether key from `roomID` can be fetched locally (by key store)
func (s *KeyService) IsLocal(roomID string) (bool, error) {
	if cache, ok := s.isLocalCache[roomID]; ok {
		if cache.expires.After(time.Now()) {
			return cache.data, nil
		}
	}

	proxy, err := s.proxy.GetRoomMasterProxy(roomID)
	if err != nil {
		return false, err
	}
	isLocal := proxy.ProxyID.Hex() == s.clientID
	s.isLocalCache[roomID] = isLocalEntry{
		data:    isLocal,
		expires: time.Now().Add(1 * time.Minute),
	}
	return isLocal, nil
}
