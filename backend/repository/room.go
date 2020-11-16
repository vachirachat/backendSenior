package repository

import (
	"backendSenior/model"
	"backendSenior/utills"
	"log"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

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

func (roomMongo RoomRepositoryMongo) GetRoomByID(roomID bson.ObjectId) (model.Room, error) {
	var room model.Room
	err := roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(roomID).One(&room)
	return room, err
}
func (roomMongo RoomRepositoryMongo) GetLastRoom() (model.Room, error) {
	var room model.Room
	err := roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).Find(nil).Sort("-created_time").One(&room)
	return room, err
}
func (roomMongo RoomRepositoryMongo) AddRoom(room model.Room) (bson.ObjectId, error) {
	room.RoomID = bson.NewObjectId()
	return room.RoomID, roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).Insert(room)
}

func (roomMongo RoomRepositoryMongo) EditRoomName(roomID bson.ObjectId, room model.Room) error {
	// objectID := bson.ObjectIdHex(roomID)
	newName := bson.M{"$set": bson.M{"roomName": room.RoomName, "updated_time": time.Now()}}
	return roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).UpdateId(roomID, newName)
}

func (roomMongo RoomRepositoryMongo) DeleteRoomByID(roomID bson.ObjectId) error {
	//objectID := bson.ObjectIdHex(roomID)
	return roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).RemoveId(roomID)
}

func (roomMongo RoomRepositoryMongo) AddMemberToRoom(roomID bson.ObjectId, listUser []bson.ObjectId) error {
	var ConnectionDB, err = mgo.Dial(utills.MONGOENDPOINT)
	var room model.Room
	if err != nil {
		return err
	}
	err = ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(roomID).One(&room)
	log.Println(roomID)
	if err != nil {
		return err
	}
	newListUser := bson.M{"$set": bson.M{"listUser": append(room.ListUser, listUser...)}}
	log.Println(newListUser)
	ConnectionDB.DB(DBRoomName).C(RoomCollection).UpdateId(roomID, newListUser)
	for _, s := range listUser {
		var user model.User
		err = ConnectionDB.DB("User").C("UserData").FindId(s).One(&user)
		newUser := bson.M{"$set": bson.M{"room": append(user.Room, roomID)}}
		userID := user.UserID
		err = ConnectionDB.DB("User").C("UserData").UpdateId(userID, newUser)
	}
	return err
}

func (roomMongo RoomRepositoryMongo) DeleteMemberToRoom(userID bson.ObjectId, roomID bson.ObjectId) error {
	var ConnectionDB, err = mgo.Dial(utills.MONGOENDPOINT)
	if err != nil {
		return err
	}
	// for delete in room
	var room model.Room
	err = ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(roomID).One(&room)
	userIDString := userID
	NewListString := utills.RemoveFormListBson(room.ListUser, userIDString)
	newUser := bson.M{"$set": bson.M{"listUser": NewListString}}
	ConnectionDB.DB(DBRoomName).C(RoomCollection).UpdateId(roomID, newUser)
	// for delete in user
	var user model.User
	err = ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(userID).One(&user)
	roomIDString := roomID
	NewListString = utills.RemoveFormListBson(user.Room, roomIDString)
	newUser = bson.M{"$set": bson.M{"room": NewListString}}
	ConnectionDB.DB("User").C("UserData").UpdateId(userID, newUser)
	return nil
}

func AddMemberToRoom(roomID bson.ObjectId, userID bson.ObjectId) error {
	var ConnectionDB, err = mgo.Dial(utills.MONGOENDPOINT)
	var room model.Room
	if err != nil {
		return err
	}
	err = ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(roomID).One(&room)
	log.Println(roomID)
	if err != nil {
		return err
	}

	newListUser := bson.M{"$set": bson.M{"listUser": append(room.ListUser, userID)}}
	log.Println(newListUser)
	ConnectionDB.DB(DBRoomName).C(RoomCollection).UpdateId(roomID, newListUser)

	var user model.User
	err = ConnectionDB.DB("User").C("UserData").FindId(userID).One(&user)
	newUser := bson.M{"$set": bson.M{"room": append(user.Room, roomID)}}
	err = ConnectionDB.DB("User").C("UserData").UpdateId(user.UserID, newUser)

	return err
}

func DeleteMemberToRoom(roomID bson.ObjectId, userID bson.ObjectId) error {
	var ConnectionDB, err = mgo.Dial(utills.MONGOENDPOINT)
	if err != nil {
		return err
	}
	// for delete in room
	var room model.Room
	err = ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(roomID).One(&room)
	userIDString := userID
	NewListString := utills.RemoveFormListBson(room.ListUser, userIDString)
	newUser := bson.M{"$set": bson.M{"listUser": NewListString}}
	ConnectionDB.DB(DBRoomName).C(RoomCollection).UpdateId(roomID, newUser)
	// for delete in user
	var user model.User
	err = ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(userID).One(&user)
	roomIDString := roomID
	NewListString = utills.RemoveFormListBson(user.Room, roomIDString)
	newUser = bson.M{"$set": bson.M{"room": NewListString}}
	ConnectionDB.DB("User").C("UserData").UpdateId(userID, newUser)
	return nil
}
