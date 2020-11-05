package repository

import (
	"backendSenior/model"
	"backendSenior/utills"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

type RoomRepository interface {
	GetAllRoom() ([]model.Room, error)
	GetLastRoom() (model.Room, error)
	GetRoomByID(roomID string) (model.Room, error)
	AddRoom(room model.Room) error
	EditRoomName(roomID bson.ObjectId, room model.Room) error
	DeleteRoomByID(roomID bson.ObjectId) error
	AddMemberToRoom(roomID bson.ObjectId, listUser []string) error
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

func (roomMongo RoomRepositoryMongo) EditRoomName(roomID bson.ObjectId, room model.Room) error {
	// objectID := bson.ObjectIdHex(roomID)
	newName := bson.M{"$set": bson.M{"room_name": room.RoomName, "updated_time": time.Now()}}
	return roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).UpdateId(roomID, newName)
}

// for add user to room userList
// func (roomMongo RoomRepositoryMongo) AddUserToRoom(roomID string, userID string) error {

// }

func (roomMongo RoomRepositoryMongo) DeleteRoomByID(roomID bson.ObjectId) error {
	//objectID := bson.ObjectIdHex(roomID)
	return roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).RemoveId(roomID)
}

func (roomMongo RoomRepositoryMongo) AddMemberToRoom(roomID bson.ObjectId, listUser []string) error {
	var ConnectionDB, err = mgo.Dial(utills.MONGOENDPOINT)
	var room model.Room
	if err != nil {
		return err
	}
	err = ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(roomID).One(&room)
	if err != nil {
		return err
	}
	newListUser := bson.M{"$set": bson.M{"listUser": append(room.ListUser, listUser...)}}
	ConnectionDB.DB(DBRoomName).C(RoomCollection).UpdateId(roomID, newListUser)
	for _, s := range listUser {
		var user model.User
		ObjectID := bson.ObjectId(s)
		err = ConnectionDB.DB("User").C("UserDate").FindId(ObjectID).One(&user)
		stringRoomID := roomID.String()
		newUser := bson.M{"$set": bson.M{"Room": append(user.Room, stringRoomID)}}
		ConnectionDB.DB("User").C("UserData").UpdateId(user.UserID, newUser)
	}
	return nil
}

// func (roomMongo RoomRepositoryMongo) DeleteMemberFromRoom(userID bson.ObjectId, roomID bson.ObjectId) error {
// 	var user model.user
// 	err :=
// }
