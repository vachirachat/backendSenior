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

var _ repository.RoomUserRepository = (*CachedRoomUserRepository)(nil)

// GetUserRooms get
func (repo *CachedRoomProxyRepository) GetUserRooms(proxyID string) (roomIDs []string, err error) {
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

// GetRoomUsers get proxyIds for specified room
func (repo *CachedRoomProxyRepository) GetRoomUsers(roomID string) (proxyIDs []string, err error) {
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

// AddUsersToRoom adds proxies to member of room, and add room to proxy's room list
// It returns error if any of proxyIDs is invalid
func (repo *CachedRoomProxyRepository) AddUsersToRoom(roomID string, proxyIDs []string) (err error) {
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

// RemoveUsersFromRoom remove proxyIDs from room in DB and cache
// return error if any of proxyIDs is invalid
func (repo *CachedRoomProxyRepository) RemoveUsersFromRoom(roomID string, proxyIDs []string) (err error) {
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
