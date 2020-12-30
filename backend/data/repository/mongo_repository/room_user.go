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

// CachedRoomUserRepository	is repository for room/user relation, with cached GET
type CachedRoomUserRepository struct {
	connection  *mgo.Session
	txnRunner   *txn.Runner
	userToRooms map[string][]string
	roomToUsers map[string][]string
	lock        sync.RWMutex
}

// NewCachedRoomUserRepository create new room user repository from mongo connection, cache isn't initialized
// Note that consistency isn't geauranteed (there might be race condition)
func NewCachedRoomUserRepository(conn *mgo.Session) *CachedRoomUserRepository {
	runner := txn.NewRunner(conn.DB(dbName).C(collectionTXNRoomUser))
	return &CachedRoomUserRepository{
		connection:  conn,
		txnRunner:   runner,
		userToRooms: make(map[string][]string),
		roomToUsers: make(map[string][]string),
		lock:        sync.RWMutex{},
	}
}

var _ repository.RoomUserRepository = (*CachedRoomUserRepository)(nil)

// GetUserRooms get RoomIDs of specified UserID
func (repo *CachedRoomUserRepository) GetUserRooms(userID string) (roomIDs []string, err error) {
	repo.lock.RLock()
	rooms, exists := repo.userToRooms[userID]
	repo.lock.RUnlock()

	if !exists {

		var user model.User
		err := repo.connection.DB(dbName).C(collectionUser).FindId(bson.ObjectIdHex(userID)).One(&user)
		if err != nil {
			return nil, err
		}

		repo.lock.Lock()
		defer repo.lock.Unlock()

		repo.userToRooms[userID] = utills.ToStringArr(user.Room)
		return repo.userToRooms[userID], nil
	}
	return rooms, nil
}

// GetRoomUsers get UserIDs of specified RoomID
func (repo *CachedRoomUserRepository) GetRoomUsers(roomID string) (userIDs []string, err error) {
	repo.lock.RLock()
	users, exist := repo.roomToUsers[roomID]
	repo.lock.RUnlock()

	if !exist {
		var room model.Room
		err := repo.connection.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).One(&room)
		if err != nil {
			return nil, err
		}

		repo.lock.Lock()
		defer repo.lock.Unlock()

		repo.roomToUsers[roomID] = utills.ToStringArr(room.ListUser)
		return repo.roomToUsers[roomID], nil
	}

	return users, nil
}

// AddUsersToRoom adds users to member of room, and add room to user's room list
// It returns error if any of userIDs is invalid
func (repo *CachedRoomUserRepository) AddUsersToRoom(roomID string, userIDs []string) (err error) {
	// Preconfition check
	n, err := repo.connection.DB(dbName).C(collectionUser).FindId(bson.M{"$in": utills.ToObjectIdArr(userIDs)}).Count()

	if err != nil {
		return err
	}
	if n != len(userIDs) {
		return fmt.Errorf("Invalid userIDs, some of them not exists %d/%d", n, len(userIDs))
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
				"$addToSet": bson.M{
					"users": bson.M{
						"$each": utills.ToObjectIdArr(userIDs),
					},
				},
			},
		},
	}
	for _, userID := range userIDs {
		ops = append(ops, txn.Op{
			C:  collectionProxy,
			Id: bson.ObjectIdHex(userID),
			Update: bson.M{
				"$addToSet": bson.M{
					"room": bson.ObjectIdHex(roomID),
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
	for _, uid := range userIDs {
		delete(repo.userToRooms, uid)
	}

	delete(repo.roomToUsers, roomID)

	return nil
}

// RemoveUsersFromRoom remove userIds from room in DB and cache
// return error if any of userIDs is invalid
func (repo *CachedRoomUserRepository) RemoveUsersFromRoom(roomID string, userIDs []string) (err error) {
	// Precondition check
	n, err := repo.connection.DB(dbName).C(collectionUser).FindId(bson.M{"$in": utills.ToObjectIdArr(userIDs)}).Count()
	if err != nil {
		return err
	}
	if n != len(userIDs) {
		return errors.New("Invalid userIDs, some of them not exists")
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
				"$pullAll": bson.M{
					"users": utills.ToObjectIdArr(userIDs),
				},
			},
		},
	}
	for _, userID := range userIDs {
		ops = append(ops, txn.Op{
			C:  collectionProxy,
			Id: bson.ObjectIdHex(userID),
			Update: bson.M{
				"$pull": bson.M{
					"room": bson.ObjectIdHex(roomID),
				},
			},
		})
	}

	err = repo.txnRunner.Run(ops, "", nil)
	if err != nil {
		return err
	}

	repo.lock.Lock()
	defer repo.lock.Unlock()

	// Invalidate cache
	for _, uid := range userIDs {
		delete(repo.userToRooms, uid)
	}

	delete(repo.roomToUsers, roomID)

	return nil
}
