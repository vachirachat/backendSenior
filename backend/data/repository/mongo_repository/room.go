package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

type RoomRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

var _ repository.RoomRepository = (*RoomRepositoryMongo)(nil)

func (roomMongo RoomRepositoryMongo) GetAllRooms() ([]model.Room, error) {
	var rooms []model.Room
	err := roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).Find(nil).All(&rooms)
	return rooms, err
}

func (roomMongo RoomRepositoryMongo) GetRoomByID(roomID string) (model.Room, error) {
	var room model.Room
	err := roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).One(&room)
	return room, err
}

func (roomMongo RoomRepositoryMongo) AddRoom(room model.Room) (string, error) {
	room.RoomID = bson.NewObjectId()
	room.ListUser = []bson.ObjectId{}
	room.ListProxy = []bson.ObjectId{}
	return room.RoomID.Hex(), roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).Insert(room)
}

// UpdateRoom updates room, return error when not found
func (roomMongo RoomRepositoryMongo) UpdateRoom(roomID string, room model.Room) error {
	updateMap := room.Map()

	delete(updateMap, "_id")
	delete(updateMap, "listUser")
	delete(updateMap, "listProxy")
	updateMap["updatedTime"] = time.Now()

	return roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).UpdateId(bson.ObjectIdHex(roomID), bson.M{
		"$set": updateMap,
	})
}

// DeleteRoomByID deletes room, return error when not found
func (roomMongo RoomRepositoryMongo) DeleteRoomByID(roomID string) error {
	objectID := bson.ObjectIdHex(roomID)
	return roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).RemoveId(objectID)
}
