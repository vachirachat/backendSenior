package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/utills"
	"errors"
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// CachedRoomProxyRepository is repository for room/proxy relation, with cached GET
type CachedRoomProxyRepository struct {
	connection    *mgo.Session
	proxyToRooms  map[string][]string
	roomToProxies map[string][]string
}

// NewCachedRoomProxyRepository create new room user repository from mongo connection, cache isn't initialized
// Note that consistency isn't geauranteed (there might be race condition)
func NewCachedRoomProxyRepository(conn *mgo.Session) *CachedRoomProxyRepository {
	return &CachedRoomProxyRepository{
		connection:    conn,
		proxyToRooms:  make(map[string][]string),
		roomToProxies: make(map[string][]string),
	}
}

var _ repository.RoomUserRepository = (*CachedRoomUserRepository)(nil)

// TODO: prevent race condition

// GetUserRooms get
func (repo *CachedRoomProxyRepository) GetUserRooms(proxyID string) (roomIDs []string, err error) {
	rooms, exists := repo.proxyToRooms[proxyID]
	if !exists {
		var proxy model.Proxy
		err := repo.connection.DB(dbName).C(collectionProxy).FindId(bson.ObjectIdHex(proxyID)).One(&proxy)
		if err != nil {
			return nil, err
		}
		repo.proxyToRooms[proxyID] = utills.ToStringArr(proxy.Rooms)
		return repo.proxyToRooms[proxyID], nil
	}
	return rooms, nil
}

// GetRoomUsers get proxyIds for specified room
func (repo *CachedRoomProxyRepository) GetRoomUsers(roomID string) (proxyIDs []string, err error) {
	proxies, exist := repo.roomToProxies[roomID]
	if !exist {
		var room model.Room
		err := repo.connection.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).One(&room)
		if err != nil {
			return nil, err
		}
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

	// Update database
	err = repo.connection.DB(dbName).C(collectionRoom).UpdateId(bson.ObjectIdHex(roomID), bson.M{
		"$addToSet": bson.M{
			"listProxy": bson.M{
				"$each": utills.ToObjectIdArr(proxyIDs), // add all from listUser to array
			},
		},
	})
	if err != nil {
		return err
	}

	err = repo.connection.DB(dbName).C(collectionProxy).UpdateId(bson.M{"$in": utills.ToObjectIdArr(proxyIDs)}, bson.M{
		"$addToSet": bson.M{
			"rooms": bson.ObjectIdHex(roomID),
		},
	})
	if err != nil {
		// TODO it should revert
		return err
	}
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

	// Update database
	err = repo.connection.DB(dbName).C(collectionRoom).UpdateId(bson.ObjectIdHex(roomID), bson.M{
		"$pullAll": bson.M{
			"listProxy": utills.ToObjectIdArr(proxyIDs),
		},
	})
	if err != nil {
		return err
	}

	err = repo.connection.DB(dbName).C(collectionProxy).UpdateId(bson.M{"$in": utills.ToObjectIdArr(proxyIDs)}, bson.M{
		"$pull": bson.M{
			"rooms": bson.ObjectIdHex(roomID),
		},
	})
	if err != nil {
		// TODO it should revert
		return err
	}
	// Invalidate cache
	for _, pid := range proxyIDs {
		delete(repo.proxyToRooms, pid)
	}

	delete(repo.roomToProxies, roomID)

	return nil
}
