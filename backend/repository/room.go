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
	DeleteRoomByID(roomID string) error
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
	err := roomMongo.ConnectionDB.DB(DBName).C(collection).Find(nil).All(&rooms)
	return rooms, err
}

func (roomMongo RoomRepositoryMongo) GetRoomByID(roomID string) (model.Room, error) {
	var room model.Room
	objectID := bson.ObjectIdHex(roomID)
	err := roomMongo.ConnectionDB.DB(DBName).C(collection).FindId(objectID).One(&room)
	return room, err
}
func (roomMongo RoomRepositoryMongo) GetLastRoom() (model.Room, error) {
	var room model.Room
	err := roomMongo.ConnectionDB.DB(DBName).C(collection).Find(nil).Sort("-created_time").One(&room)
	return room, err
}
func (roomMongo RoomRepositoryMongo) AddRoom(room model.Room) error {
	return roomMongo.ConnectionDB.DB(DBName).C(collection).Insert(room)
}

func (roomMongo RoomRepositoryMongo) EditRoomName(roomID string, room model.Room) error {
	objectID := bson.ObjectIdHex(roomID)
	newName := bson.M{"$set": bson.M{"room_name": room.RoomName, "updated_time": time.Now()}}
	return roomMongo.ConnectionDB.DB(DBName).C(collection).UpdateId(objectID, newName)
}

func (roomMongo RoomRepositoryMongo) DeleteRoomByID(roomID string) error {
	objectID := bson.ObjectIdHex(roomID)
	return roomMongo.ConnectionDB.DB(DBName).C(collection).RemoveId(objectID)
}
