package key_service

// manage local key

import (
	"errors"
	"fmt"
	model_proxy "proxySenior/domain/model"
	"proxySenior/utils"

	"github.com/globalsign/mgo/bson"
)

// copyKeysSlice return copy of keys slice
func copyKeysSlice(keys []model_proxy.KeyRecord) []model_proxy.KeyRecord {
	cpy := make([]model_proxy.KeyRecord, len(keys))
	copy(cpy, keys)
	return cpy
}

// GetKeyLocal is used for getting keys for room locally if possible
func (s *KeyService) GetKeyLocal(roomID string) ([]model_proxy.KeyRecord, error) {
	ok, err := s.IsLocal(roomID)
	if err != nil {
		return nil, fmt.Errorf("error checking locality of key: %v", err)
	}

	if !ok {
		return nil, errors.New("can't get local key for remote room")
	}

	if keys, ok := s.keyCache[roomID]; ok {
		return copyKeysSlice(keys), nil
	}

	keys, err := s.local.Find(model_proxy.KeyRecordUpdate{
		RoomID: bson.ObjectIdHex(roomID),
	})

	if err == nil && keys == nil {
		keys = []model_proxy.KeyRecord{}
	}

	s.keyCache[roomID] = keys
	return copyKeysSlice(keys), nil
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
	// invalidate self key
	if err != nil {
		s.InvalidateKeyCache(roomID)
	}

	return err
}

// IsLocal return whether key from `roomID` can be fetched locally (by key store)
func (s *KeyService) IsLocal(roomID string) (bool, error) {
	if proxy, ok := s.roomMasterCache[roomID]; ok {
		return proxy.ProxyID.Hex() == utils.ClientID, nil
	}

	proxy, err := s.proxy.GetRoomMasterProxy(roomID)
	if err != nil {
		return false, err
	}
	isLocal := proxy.ProxyID.Hex() == s.clientID
	s.roomMasterCache[roomID] = proxy
	return isLocal, nil
}

// InvalidateRoomMaster invalidate cached query of locality
func (s *KeyService) InvalidateRoomMaster(roomID string) {
	delete(s.roomMasterCache, roomID)
}

// InvalidateKeyCache invalidate cached key (forcing to get new key)
func (s *KeyService) InvalidateKeyCache(roomID string) {
	delete(s.keyCache, roomID)
}
