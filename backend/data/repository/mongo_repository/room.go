package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/utills"
	"log"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

type RoomRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

var _ repository.RoomRepository = (*RoomRepositoryMongo)(nil)

// สำหรับยัด UserID
func toObjectIdArr(stringArr []string) []bson.ObjectId {
	result := make([]bson.ObjectId, len(stringArr))
	n := len(stringArr)
	for i := 0; i < n; i++ {
		result[i] = bson.ObjectIdHex(stringArr[i])
	}
	return result
}

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
	return room.RoomID.Hex(), roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).Insert(room)
}

// UpdateRoom updates room, return error when not found
func (roomMongo RoomRepositoryMongo) UpdateRoom(roomID string, room model.Room) error {
	updateMap := room.Map()

	// delete(updateMap, "_id")
	// delete(updateMap, "listUser")
	updateMap["updatedTime"] = time.Now()

	return roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).UpdateId(bson.ObjectIdHex(roomID), bson.M{
		"$set": updateMap,
	})
}

// DeleteRoomByID deletes room, return error when not found
func (roomMongo RoomRepositoryMongo) DeleteRoomByID(roomID string) error {
	//objectID := stringHex(roomID)
	return roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).RemoveId(bson.ObjectIdHex(roomID))
}

// AddMemberToRoom appends member ids to room, return errors when it doesn't exists
func (roomMongo RoomRepositoryMongo) AddMemberToRoom(roomID string, listUser []string) error {
	// TODO might need to fix logic
	log.Println("this is roomID")
	log.Println(roomID)
	err := roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).UpdateId(bson.ObjectIdHex(roomID), bson.M{
		"$push": bson.M{
			"listUser": bson.M{
				"$each": listUser, // add all from listUser to array
			},
		},
	})
	if err != nil {
		return err
	}

	var room model.Room
	err = roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).One(&room)
	if err != nil {
		return err
	}
	for _, s := range listUser {
		var user model.User
		err = roomMongo.ConnectionDB.DB(collectionUser).C(collectionUser).FindId(bson.ObjectIdHex(s)).One(&user)
		didntAdd := false
		for _, v := range user.Room {
			if v == bson.ObjectIdHex(roomID) {
				didntAdd = true
			}
		}
		if didntAdd == false {
			newUser := bson.M{"$set": bson.M{"room": append(user.Room, bson.ObjectIdHex(roomID))}}
			userID := bson.ObjectIdHex(s)
			err = roomMongo.ConnectionDB.DB(collectionUser).C(collectionUser).UpdateId(userID, newUser)
		}

	}

	return err
}

func (roomMongo RoomRepositoryMongo) DeleteMemberFromRoom(roomID string, userID []string) error {
	// TODO might need to fix logic
	var ConnectionDB, err = mgo.Dial(utills.MONGOENDPOINT)
	if err != nil {
		return err
	}
	// for delete in room
	var room model.Room
	err = ConnectionDB.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).One(&room)

	var NewListString []bson.ObjectId
	// TODO fix this, i just want log.Println for fix error didnt use
	for _, v := range toObjectIdArr(userID) {
		NewListString := utills.RemoveFormListBson(room.ListUser, v)
		log.Println(NewListString)
	}
	newUser := bson.M{"$set": bson.M{"listUser": NewListString}}
	ConnectionDB.DB(dbName).C(collectionRoom).UpdateId(bson.ObjectIdHex(roomID), newUser)
	// for delete in user
	// var user model.User
	// err = ConnectionDB.DB(collectionUser).C(collectionUser).FindId(userID).One(&user)
	// roomIDString := roomID
	for _, v := range toObjectIdArr(userID) {
		var user model.User
		err = ConnectionDB.DB(collectionUser).C(collectionUser).FindId(v).One(&user)
		NewListString = utills.RemoveFormListBson(user.Room, bson.ObjectIdHex(roomID))
		newUser = bson.M{"$set": bson.M{"room": NewListString}}
		ConnectionDB.DB(collectionUser).C(collectionUser).UpdateId(v, newUser)
	}

	return nil
}
