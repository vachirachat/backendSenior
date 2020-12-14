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

func toStringArr(objIdArr []bson.ObjectId) []string {
	var result = make([]string, len(objIdArr))
	n := len(objIdArr)
	for i := 0; i < n; i++ {
		result[i] = objIdArr[i].Hex()
		fmt.Printf("obj %s hex %s\n", objIdArr[i], objIdArr[i].Hex())
	}
	return result
}

// CachedRoomUserRepository	is repository for room/user relation, with cached GET
type CachedRoomUserRepository struct {
	connection  *mgo.Session
	userToRooms map[string][]string
	roomToUsers map[string][]string
}

// NewCachedRoomUserRepository create new room user repository from mongo connection, cache isn't initialized
// Note that consistency isn't geauranteed (there might be race condition)
func NewCachedRoomUserRepository(conn *mgo.Session) *CachedRoomUserRepository {
	return &CachedRoomUserRepository{
		connection:  conn,
		userToRooms: make(map[string][]string),
		roomToUsers: make(map[string][]string),
	}
}

var _ repository.RoomUserRepository = (*CachedRoomUserRepository)(nil)

// TODO: prevent race condition

// GetUserRooms get RoomIDs of specified UserID
func (repo *CachedRoomUserRepository) GetUserRooms(userID string) (roomIDs []string, err error) {
	rooms, exists := repo.userToRooms[userID]
	if !exists {
		var user model.User
		err := repo.connection.DB(dbName).C(collectionUser).FindId(bson.ObjectIdHex(userID)).One(&user)
		if err != nil {
			return nil, err
		}
		repo.userToRooms[userID] = toStringArr(user.Room)
		return repo.userToRooms[userID], nil
	}
	return rooms, nil
}

// GetRoomUsers get UserIDs of specified RoomID
func (repo *CachedRoomUserRepository) GetRoomUsers(roomID string) (userIDs []string, err error) {
	users, exist := repo.roomToUsers[roomID]
	if !exist {
		var room model.Room
		err := repo.connection.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).One(&room)
		if err != nil {
			return nil, err
		}
		repo.roomToUsers[roomID] = toStringArr(room.ListUser)
		return repo.roomToUsers[roomID], nil
	}
	return users, nil
}

// AddUsersToRoom adds users to member of room, and add room to user's room list
// It returns error if any of userIDs is invalid
func (repo *CachedRoomUserRepository) AddUsersToRoom(roomID string, userIDs []string) (err error) {
	// Preconfition check
	n, err := repo.connection.DB(dbName).C(collectionUser).FindId(toObjectIdArr(userIDs)).Count()
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

	// Update database
	err = repo.connection.DB(dbName).C(collectionRoom).UpdateId(roomID, bson.M{
		"$push": bson.M{
			"listUser": bson.M{
				"$each": userIDs, // add all from listUser to array
			},
		},
	})
	if err != nil {
		return err
	}

	err = repo.connection.DB(dbName).C(collectionUser).UpdateId(userIDs, bson.M{
		"$push": bson.M{
			"room": roomID,
		},
	})
	if err != nil {
		// TODO it should revert
		return err
	}
	// Update cache
	for _, uid := range userIDs {
		repo.userToRooms[uid] = append(repo.userToRooms[uid], roomID)
	}

	repo.roomToUsers[roomID] = append(repo.roomToUsers[roomID], userIDs...)

	return nil
}

// RemoveUsersFromRoom remove userIds from room in DB and cache
// return error if any of userIDs is invalid
func (repo *CachedRoomUserRepository) RemoveUsersFromRoom(roomID string, userIDs []string) (err error) {
	// Precondition check
	n, err := repo.connection.DB(dbName).C(collectionUser).FindId(toObjectIdArr(userIDs)).Count()
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

	// Update database
	err = repo.connection.DB(dbName).C(collectionRoom).UpdateId(roomID, bson.M{
		"$pullAll": bson.M{
			"listUser": userIDs,
		},
	})
	if err != nil {
		return err
	}

	err = repo.connection.DB(dbName).C(collectionUser).UpdateId(userIDs, bson.M{
		"$pull": bson.M{
			"room": roomID,
		},
	})
	if err != nil {
		// TODO it should revert
		return err
	}
	// Update cache
	for _, uid := range userIDs {
		repo.userToRooms[uid], _ = utills.ArrStringRemoveMatched(repo.userToRooms[uid], []string{roomID})
	}

	repo.roomToUsers[roomID], _ = utills.ArrStringRemoveMatched(repo.roomToUsers[roomID], userIDs)

	return nil
}
