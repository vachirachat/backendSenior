package key_service

// manage local key

import (
	"backendSenior/domain/model"
	"errors"
	"fmt"
	"proxySenior/config"
	model_proxy "proxySenior/domain/model"
	"time"

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

	if _keys, ok := s.keyCache.Get(roomID); ok {
		keys := _keys.([]model_proxy.KeyRecord)
		return copyKeysSlice(keys), nil
	}

	keys, err := s.local.Find(model_proxy.KeyRecordUpdate{
		RoomID: bson.ObjectIdHex(roomID),
	})

	if err == nil && keys == nil {
		keys = []model_proxy.KeyRecord{}
	}

	s.keyCache.Set(roomID, keys)
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
		s.RevalidateKeyCache(roomID)
	}

	return err
}

// IsLocal return whether key from `roomID` can be fetched locally (by key store)
func (s *KeyService) IsLocal(roomID string) (bool, error) {
	if _proxy, ok := s.roomMasterCache.Get(roomID); ok {
		proxy := _proxy.(model.Proxy)
		return proxy.ProxyID.Hex() == config.ClientID, nil
	}

	proxy, err := s.proxy.GetRoomMasterProxy(roomID)
	if err != nil {
		return false, err
	}
	isLocal := proxy.ProxyID.Hex() == s.clientID
	s.roomMasterCache.Set(roomID, proxy)
	return isLocal, nil
}

// RevalidateRoomMaster revalidate cached query of locality
func (s *KeyService) RevalidateRoomMaster(roomID string) {
	s.roomMasterCache.Del(roomID)
	time.Sleep(100 * time.Millisecond) // have some buffer time
	_, _ = s.IsLocal(roomID)
}

// RevalidateKeyCache revalidate cached key (delete and get key again)
func (s *KeyService) RevalidateKeyCache(roomID string) {
	s.keyCache.Del(roomID)
	time.Sleep(100 * time.Millisecond)
	s._ensureKey(roomID)
}

func (s *KeyService) _ensureKey(roomID string) {
	if local, err := s.IsLocal(roomID); err != nil {
		fmt.Println("[warn] invalidate all: isLocal", roomID, ":", err)
	} else if local {
		_, _ = s.GetKeyLocal(roomID)
	} else if !local {
		_, _ = s.GetKeyRemote(roomID)
	}
}

// RevalidateAll revalidate locality, and keys of all rooms.
// additionally, it also get key of all rooms that proxy is in
func (s *KeyService) RevalidateAll() {
	fmt.Println("invalidating and re-getting all keys")
	for kv := range s.roomMasterCache.Iter() {
		go s.RevalidateRoomMaster(kv.Key.(string))
	}
	for kv := range s.keyCache.Iter() {
		go s.RevalidateKeyCache(kv.Key.(string))
	}

	proxy, err := s.proxy.GetProxyByID(config.ClientID)
	if err != nil {
		fmt.Printf("revalidate all: %v\n", err)
	} else {
		for _, roomID := range proxy.Rooms {
			s._ensureKey(roomID.Hex())
			_, _ = s.IsLocal(roomID.Hex())
		}
	}
}
