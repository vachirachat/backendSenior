package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/utills"
	"errors"
	"fmt"
	"sync"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/globalsign/mgo/txn"
)

// CachedRoomProxyRepository is repository for room/proxy relation, with cached GET
type CachedRoomProxyRepository struct {
	connection    *mgo.Session
	txnRunner     *txn.Runner
	proxyToRooms  map[string][]string
	roomToProxies map[string][]string
	lock          sync.RWMutex
}

// NewCachedRoomProxyRepository create new room user repository from mongo connection, cache isn't initialized
// Note that consistency isn't geauranteed (there might be race condition)
func NewCachedRoomProxyRepository(conn *mgo.Session) *CachedRoomProxyRepository {
	runner := txn.NewRunner(conn.DB(dbName).C(collectionTXNRoomUser))
	return &CachedRoomProxyRepository{
		connection:    conn,
		txnRunner:     runner,
		proxyToRooms:  make(map[string][]string),
		roomToProxies: make(map[string][]string),
		lock:          sync.RWMutex{},
	}
}

var _ repository.RoomProxyRepository = (*CachedRoomProxyRepository)(nil)

// GetProxyRooms get proxyIDs in room
func (repo *CachedRoomProxyRepository) GetProxyRooms(proxyID string) (roomIDs []string, err error) {
	repo.lock.RLock()
	rooms, exists := repo.proxyToRooms[proxyID]
	repo.lock.RUnlock()

	if !exists {
		var proxy model.Proxy
		err := repo.connection.DB(dbName).C(collectionProxy).FindId(bson.ObjectIdHex(proxyID)).One(&proxy)
		if err != nil {
			return nil, err
		}

		repo.lock.Lock()
		defer repo.lock.Unlock()

		repo.proxyToRooms[proxyID] = utills.ToStringArr(proxy.Rooms)
		return repo.proxyToRooms[proxyID], nil
	}
	return rooms, nil
}

// GetRoomProxies get roomIDs that proxy is in
func (repo *CachedRoomProxyRepository) GetRoomProxies(roomID string) (proxyIDs []string, err error) {
	repo.lock.RLock()
	proxies, exist := repo.roomToProxies[roomID]
	repo.lock.RUnlock()

	if !exist {
		var room model.Room
		err := repo.connection.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).One(&room)
		if err != nil {
			return nil, err
		}

		repo.lock.Lock()
		defer repo.lock.Unlock()

		repo.roomToProxies[roomID] = utills.ToStringArr(room.ListProxy)
		return repo.roomToProxies[roomID], nil
	}

	return proxies, nil
}

// AddProxiesToRoom adds proxies to room, and add room to proxy's room list
// It returns error if any of proxyIDs is invalid
func (repo *CachedRoomProxyRepository) AddProxiesToRoom(roomID string, proxyIDs []string) (err error) {
	// Preconfition check
	n, err := repo.connection.DB(dbName).C(collectionProxy).FindId(bson.M{"$in": utills.ToObjectIdArr(proxyIDs)}).Count()

	if err != nil {
		return err
	}
	if n != len(proxyIDs) {
		return fmt.Errorf("Invalid proxyIDs, some of them not exists %d/%d", n, len(proxyIDs))
	}

	n, err = repo.connection.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).Count()
	if n != 1 {
		return errors.New("Invalid Room ID")
	}

	ops := []txn.Op{
		{
			C:  collectionRoom,
			Id: bson.ObjectIdHex(roomID),
			Update: bson.M{
				"$addToSet": model.RoomUpdateMongo{
					ListProxy: bson.M{
						"$each": utills.ToObjectIdArr(proxyIDs),
					},
				},
			},
		},
	}
	for _, proxyID := range proxyIDs {
		ops = append(ops, txn.Op{
			C:  collectionProxy,
			Id: bson.ObjectIdHex(proxyID),
			Update: bson.M{
				"$addToSet": model.ProxyUpdateMongo{
					Rooms: bson.ObjectIdHex(roomID),
				},
			},
		})
	}

	err = repo.txnRunner.Run(ops, "", nil)
	if err != nil {
		return err
	}

	repo.lock.Lock()
	repo.lock.Unlock()

	// Invalidate cache
	for _, pid := range proxyIDs {
		delete(repo.proxyToRooms, pid)
	}

	delete(repo.roomToProxies, roomID)

	return nil
}

// RemoveProxiesFromRoom remove proxyIDs from room in DB and cache
// return error if any of proxyIDs is invalid
func (repo *CachedRoomProxyRepository) RemoveProxiesFromRoom(roomID string, proxyIDs []string) (err error) {
	// Precondition check
	n, err := repo.connection.DB(dbName).C(collectionProxy).FindId(bson.M{"$in": utills.ToObjectIdArr(proxyIDs)}).Count()
	if err != nil {
		return err
	}
	if n != len(proxyIDs) {
		return errors.New("Invalid proxyIDs, some of them not exists")
	}

	n, err = repo.connection.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).Count()
	if n != 1 {
		return errors.New("Invalid Room ID")
	}

	ops := []txn.Op{
		{
			C:  collectionRoom,
			Id: bson.ObjectIdHex(roomID),
			Update: bson.M{
				"$pullAll": model.RoomUpdateMongo{
					ListProxy: utills.ToObjectIdArr(proxyIDs),
				},
			},
		},
	}
	for _, proxyID := range proxyIDs {
		ops = append(ops, txn.Op{
			C:  collectionProxy,
			Id: bson.ObjectIdHex(proxyID),
			Update: bson.M{
				"$pull": model.ProxyUpdateMongo{
					Rooms: bson.ObjectIdHex(roomID),
				},
			},
		})
	}

	err = repo.txnRunner.Run(ops, "", nil)
	if err != nil {
		return err
	}

	repo.lock.Lock()
	repo.lock.Unlock()

	// Invalidate cache
	for _, pid := range proxyIDs {
		delete(repo.proxyToRooms, pid)
	}

	delete(repo.roomToProxies, roomID)

	return nil
}

// GetRoomMasterProxy return the proxyID that is master of room
func (repo *CachedRoomProxyRepository) GetRoomMasterProxy(roomID string) (model.Proxy, error) {
	var room model.Room
	err := repo.connection.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).One(&room)
	if err != nil {
		return model.Proxy{}, err
	}
	if room.MainProxy == "" {
		return model.Proxy{}, errors.New("master proxy not set")
	}

	var proxy model.Proxy
	err = repo.connection.DB(dbName).C(collectionProxy).FindId(room.MainProxy).One(&proxy)

	return proxy, err
}

// SetRoomMasterProxy change main proxy of the room
func (repo *CachedRoomProxyRepository) SetRoomMasterProxy(roomID string, mainProxyID string) error {
	err := repo.connection.DB(dbName).C(collectionRoom).UpdateId(bson.ObjectIdHex(roomID), model.RoomUpdateMongo{
		MainProxy: bson.ObjectIdHex(mainProxyID),
	})
	return err
}

// GetProxyMasterRooms get room of which proxy is master
func (repo *CachedRoomProxyRepository) GetProxyMasterRooms(proxyID string) ([]string, error) {
	var rooms []model.Room
	err := repo.connection.DB(dbName).C(collectionRoom).Find(model.RoomUpdateMongo{
		MainProxy: bson.ObjectIdHex(proxyID),
	}).All(&rooms)
	if err != nil {
		if err.Error() == "not found" {
			return []string{}, nil
		}
		return nil, err
	}
	roomIDs := make([]string, len(rooms))
	for i := range rooms {
		roomIDs[i] = rooms[i].MainProxy.Hex()
	}
	return roomIDs, nil
}
