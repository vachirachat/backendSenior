package repository

import (
	"backendSenior/domain/model"

	"github.com/globalsign/mgo/bson"
)

// RoomRepository defines interface for room repo
type RoomRepository interface {
	GetAllRoom() ([]model.Room, error)
	GetLastRoom() (model.Room, error)
	GetRoomByID(roomID bson.ObjectId) (model.Room, error)
	AddRoom(room model.Room) (bson.ObjectId, error)
	EditRoomName(roomID bson.ObjectId, room model.Room) error
	DeleteRoomByID(roomID bson.ObjectId) error
	AddMemberToRoom(roomID bson.ObjectId, listUser []bson.ObjectId) error
	DeleteMemberToRoom(userID bson.ObjectId, roomID bson.ObjectId) error
}
