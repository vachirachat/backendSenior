package service

import (
	"backendSenior/domain/model"
	"backendSenior/utills"
	"errors"
	"fmt"
	"sync"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// KeyExchangeService decide master of the room
type KeyExchangeService struct {
	col      *mgo.Collection // collection for key
	isOnline map[string]bool // online status of proxy
}

// NewKeyExchangeService create key exchange service
func NewKeyExchangeService(col *mgo.Collection) *KeyExchangeService {
	return &KeyExchangeService{
		col:      col,
		isOnline: make(map[string]bool),
	}
}

// SetOnline set online status of procy
func (s *KeyExchangeService) SetOnline(proxyID string, online bool) {
	s.isOnline[proxyID] = online
}

// CatchupKeyVersion set key version of roomId, proxyId to latest version possible
func (s *KeyExchangeService) CatchupKeyVersion(roomID string, proxyID string) error {
	var kv model.KeyVersion
	err := s.col.Find(model.KeyVersionFilter{
		RoomID: bson.ObjectIdHex(roomID),
	}).Sort("-version", "-priority", "proxyId").One(&kv)
	latest := kv.Version
	if err != nil {
		return fmt.Errorf("error finding latest version: %w", err)
	}

	err = s.col.Update(model.KeyVersionFilter{RoomID: bson.ObjectIdHex(roomID), ProxyID: bson.ObjectIdHex(proxyID)}, bson.M{
		"$set": model.KeyVersionFilter{Version: latest},
	})
	if err != nil {
		return fmt.Errorf("error updating: %w", err)
	}
	return nil
}

// IncrementVersion increase version of roomId in proxyId, it performs no check
func (s *KeyExchangeService) IncrementVersion(roomID string, proxyID string) error {
	_, err := s.col.Upsert(model.KeyVersionFilter{RoomID: bson.ObjectIdHex(roomID), ProxyID: bson.ObjectIdHex(proxyID)}, bson.M{
		"$inc": model.KeyVersionFilter{Version: 1, Priority: 0},
	})
	if err != nil {
		return fmt.Errorf("error updating: %w", err)
	}
	return nil
}

// GetMaster return master proxy of room, considering online status
func (s *KeyExchangeService) GetMaster(roomID string) (string, error) {

	var kvs []model.KeyVersion
	s.col.Find(model.KeyVersionFilter{
		RoomID: bson.ObjectIdHex(roomID),
	}).Sort("-version", "-priority", "proxyId").All(&kvs)

	if len(kvs) == 0 {
		return "", fmt.Errorf("no keyversion data in room %s, database corrupt ? or no proxy ?", roomID)
	}

	latest := kvs[0].Version
	for _, kv := range kvs {
		if !s.isOnline[kv.ProxyID.Hex()] {
			continue
		}
		if kv.Version < latest {
			return "", errors.New("latest version proxy is offline, can't proceed w/o losing consistency")
		}
		return kv.ProxyID.Hex(), nil
	}
	return "", errors.New("all proxies are offline")
}

// SetPriority create of set priority of room
func (s *KeyExchangeService) SetPriority(roomID string, proxyID string, priority int) error {
	_, err := s.col.Upsert(model.KeyVersionFilter{RoomID: bson.ObjectIdHex(roomID), ProxyID: bson.ObjectIdHex(proxyID)}, bson.M{
		"$set": model.KeyVersionFilter{Priority: priority},
		"$inc": model.KeyVersionFilter{Version: 0}, // force create field if not exists
	})
	return err
}

// GetPriorities return priorities of all proxy in room
func (s *KeyExchangeService) GetPriorities(roomID string) ([]model.KeyVersion, error) {
	var kvs []model.KeyVersion
	err := s.col.Find(model.KeyVersionFilter{RoomID: bson.ObjectIdHex(roomID)}).All(&kvs)
	return kvs, err
}

type MultipleErrors struct {
	error
	Errors []error
}

func (e MultipleErrors) Error() string {
	return fmt.Sprint(len(e.Errors), "errors occured")
}

// Ensure keyversion is available
func (s *KeyExchangeService) Ensure(roomID string, proxyIDs []string) error {
	var wg sync.WaitGroup
	errors := []error{}
	for _, pid := range proxyIDs {
		wg.Add(1)
		go func(pid string) {
			_, err := s.col.Upsert(model.KeyVersionFilter{RoomID: bson.ObjectIdHex(roomID), ProxyID: bson.ObjectIdHex(pid)}, bson.M{
				"$inc": model.KeyVersionFilter{Version: 0, Priority: 0}, // force create field if not exists
			})
			if err != nil {
				errors = append(errors, err)
			}
			wg.Done()
		}(pid)
	}
	wg.Wait()
	if len(errors) == 0 {
		return nil
	}
	return MultipleErrors{Errors: errors}
}

// Delete delete key version from database
func (s *KeyExchangeService) Delete(roomID string, proxyIDs []string) error {
	_, err := s.col.RemoveAll(model.KeyVersionFilter{
		RoomID:  roomID,
		ProxyID: bson.M{"$in": utills.ToObjectIdArr(proxyIDs)},
	})
	return err
}
