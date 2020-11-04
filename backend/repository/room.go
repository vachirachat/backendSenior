package repository

import (
	"backendSenior/model"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

type RoomRepository interface {
	GetAllRoom() ([]model.Room, error)
	GetLastRoom() (model.Room, error)
	GetRoomByID(roomID string) (model.Room, error)
	AddRoom(room model.Room) error
	EditRoomName(roomID string, room model.Room) error
	DeleteRoomByID(roomID bson.ObjectId) error
}

type RoomRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

const (
	DBRoomName     = "Room"
	RoomCollection = "RoomData"
)

func (roomMongo RoomRepositoryMongo) GetAllRoom() ([]model.Room, error) {
	var rooms []model.Room
	err := roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).Find(nil).All(&rooms)
	return rooms, err
}

func (roomMongo RoomRepositoryMongo) GetRoomByID(roomID string) (model.Room, error) {
	var room model.Room
	objectID := bson.ObjectIdHex(roomID)
	err := roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(objectID).One(&room)
	return room, err
}
func (roomMongo RoomRepositoryMongo) GetLastRoom() (model.Room, error) {
	var room model.Room
	err := roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).Find(nil).Sort("-created_time").One(&room)
	return room, err
}
func (roomMongo RoomRepositoryMongo) AddRoom(room model.Room) error {
	return roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).Insert(room)
}

func (roomMongo RoomRepositoryMongo) EditRoomName(roomID string, room model.Room) error {
	objectID := bson.ObjectIdHex(roomID)
	newName := bson.M{"$set": bson.M{"room_name": room.RoomName, "updated_time": time.Now()}}
	return roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).UpdateId(objectID, newName)
}

// for add user to room userList
// func (roomMongo RoomRepositoryMongo) AddUserToRoom(roomID string, userID string) error {

// }

func (roomMongo RoomRepositoryMongo) DeleteRoomByID(roomID bson.ObjectId) error {
	//objectID := bson.ObjectIdHex(roomID)
	return roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).RemoveId(roomID)
}
