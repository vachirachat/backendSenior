package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/utills"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

type RoomRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

const (
	DBRoomName     = "mychat"
	RoomCollection = "rooms"
)

var _ repository.RoomRepository = (*RoomRepositoryMongo)(nil)

func toObjectIdArr(stringArr []string) []bson.ObjectId {
	result := make([]bson.ObjectId, len(stringArr))
	n := len(stringArr)
	for i := 0; i < n; i++ {
		result[i] = bson.ObjectId(stringArr[i])
	}
	return result
}

func (roomMongo RoomRepositoryMongo) GetAllRooms() ([]model.Room, error) {
	var rooms []model.Room
	err := roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).Find(nil).All(&rooms)
	return rooms, err
}

func (roomMongo RoomRepositoryMongo) GetRoomByID(roomID string) (model.Room, error) {
	var room model.Room
	err := roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(roomID).One(&room)
	return room, err
}

func (roomMongo RoomRepositoryMongo) AddRoom(room model.Room) (string, error) {
	room.RoomID = bson.NewObjectId()
	return room.RoomID.Hex(), roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).Insert(room)
}

// UpdateRoom updates room, return error when not found
func (roomMongo RoomRepositoryMongo) UpdateRoom(roomID string, room model.Room) error {
	updateMap := room.Map()

	delete(updateMap, "_id")
	delete(updateMap, "listUser")
	updateMap["updatedTime"] = time.Now()

	return roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).UpdateId(roomID, bson.M{
		"$set": updateMap,
	})
}

// DeleteRoomByID deletes room, return error when not found
func (roomMongo RoomRepositoryMongo) DeleteRoomByID(roomID string) error {
	//objectID := stringHex(roomID)
	return roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).RemoveId(roomID)
}

// AddMemberToRoom appends member ids to room, return errors when it doesn't exists
func (roomMongo RoomRepositoryMongo) AddMemberToRoom(roomID string, listUser []string) error {
	// TODO might need to fix logic
	err := roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).UpdateId(roomID, bson.M{
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
	err = roomMongo.ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(roomID).One(&room)
	if err != nil {
		return err
	}

	for _, s := range room.ListUser {
		var user model.User
		err = roomMongo.ConnectionDB.DB("User").C("UserData").FindId(s).One(&user)
		newUser := bson.M{"$set": bson.M{"room": append(user.Room, bson.ObjectId(roomID))}}
		userID := user.UserID
		err = roomMongo.ConnectionDB.DB("User").C("UserData").UpdateId(userID, newUser)
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
	err = ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(roomID).One(&room)

	// TODO fix this, i just want it to compile for now
	NewListString := utills.RemoveFormListBson(room.ListUser, toObjectIdArr(userID)[0])
	newUser := bson.M{"$set": bson.M{"listUser": NewListString}}
	ConnectionDB.DB(DBRoomName).C(RoomCollection).UpdateId(roomID, newUser)
	// for delete in user
	var user model.User
	err = ConnectionDB.DB(DBRoomName).C(RoomCollection).FindId(userID).One(&user)
	roomIDString := roomID
	NewListString = utills.RemoveFormListBson(user.Room, bson.ObjectId(roomIDString))
	newUser = bson.M{"$set": bson.M{"room": NewListString}}
	ConnectionDB.DB("User").C("UserData").UpdateId(userID, newUser)
	return nil
}
