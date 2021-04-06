package repository

import "github.com/globalsign/mgo/bson"

// RoomUserRepository is interface for repository managing room/user relation
type RoomUserRepository interface {
	GetUserRooms(userID string) (roomIDs []string, err error)
	GetRoomUsers(roomID string) (userIDs []string, err error)
	AddUsersToRoom(roomID string, userIDs []string) (err error)
	RemoveUsersFromRoom(roomID string, userIDs []string) (err error)

	// AddAdmins should invite user and promote them to admin
	AddAdminsToRoom(roomID bson.ObjectId, userIDs []bson.ObjectId) (err error)
	// RemoveAdminsFromRoom should only demote admin (if want to kick user remove user)
	RemoveAdminsFromRoom(roomID bson.ObjectId, userIDs []bson.ObjectId) (err error)
}
