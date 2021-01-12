package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"proxySenior/domain/interface/repository"
	model_proxy "proxySenior/domain/model"
	"time"

	"github.com/globalsign/mgo/bson"
)

// KeyService is service for managing key
type KeyService struct {
	local    repository.Keystore // local
	remote   repository.RemoteKeyStore
	proxy    repository.ProxyMasterAPI
	clientID string // clientID of this proxy
}

// NewKeyService create new key service
func NewKeyService(local repository.Keystore, remote repository.RemoteKeyStore, proxy repository.ProxyMasterAPI, clientID string) *KeyService {
	return &KeyService{
		local:    local,
		remote:   remote,
		proxy:    proxy,
		clientID: clientID,
	}
}

// ---- local

// GetKeyLocal is used for getting keys for room locally if possible
func (s *KeyService) GetKeyLocal(roomID string) ([]model_proxy.KeyRecord, error) {
	ok, err := s.IsLocal(roomID)
	if err != nil {
		return nil, fmt.Errorf("error checking locality of key: %v", err)
	}

	if ok {
		keys, err := s.local.Find(model_proxy.KeyRecordUpdate{
			RoomID: bson.ObjectIdHex(roomID),
		})
		if err == nil && keys == nil {
			keys = []model_proxy.KeyRecord{}
		}
		return keys, err
	}

	return nil, errors.New("can't get local key for remote room")
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

type isLocalEntry struct {
	data    bool
	expires time.Time
}

var isLocalCache = make(map[string]isLocalEntry)

// IsLocal return whether key from `roomID` can be fetched locally (by key store)
func (s *KeyService) IsLocal(roomID string) (bool, error) {
	if cache, ok := isLocalCache[roomID]; ok {
		if cache.expires.After(time.Now()) {
			return cache.data, nil
		}
	}

	proxy, err := s.proxy.GetRoomMasterProxy(roomID)
	if err != nil {
		return false, err
	}
	isLocal := proxy.ProxyID.Hex() == s.clientID
	isLocalCache[roomID] = isLocalEntry{
		data:    isLocal,
		expires: time.Now().Add(1 * time.Minute),
	}
	return isLocal, nil
}

type keyRemoteEntry struct {
	data    []model_proxy.KeyRecord
	expires time.Time
}

var keyRemoteCache = make(map[string]keyRemoteEntry)

// GetKeyRemote is used to get key from remote
func (s *KeyService) GetKeyRemote(roomID string) ([]model_proxy.KeyRecord, error) {
	if cache, ok := keyRemoteCache[roomID]; ok {
		if cache.expires.After(time.Now()) {
			return cache.data, nil
		}
	}

	rec, err := s.remote.GetByRoom(roomID)
	if err != nil {
		return nil, err
	}

	keyRemoteCache[roomID] = keyRemoteEntry{
		data:    rec,
		expires: time.Now().Add(10 * time.Second),
	}

	return rec, nil
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
