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

// // AddMemberToRoom appends member ids to room, return errors when it doesn't exists
// func (roomMongo RoomRepositoryMongo) AddMemberToRoom(roomID string, listUser []string) error {
// 	// TODO might need to fix logic
// 	bid := bson.ObjectIdHex(roomID)
// 	err := roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).UpdateId(bid, bson.M{
// 		"$push": bson.M{
// 			"listUser": bson.M{
// 				"$each": utills.ToObjectIdArr(listUser), // add all from listUser to array
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return fmt.Errorf("error updating room users %s", err)
// 	}

// 	var room model.Room
// 	err = roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).One(&room)
// 	if err != nil {
// 		return fmt.Errorf("error re-getting room users %s", err)
// 	}

// 	for _, s := range room.ListUser {
// 		var user model.User
// 		err = roomMongo.ConnectionDB.DB(dbName).C(collectionUser).FindId(s).One(&user)
// 		if err != nil {
// 			err = fmt.Errorf("error finding user %s", err)
// 			break
// 		}
// 		newUser := bson.M{"$set": bson.M{"room": append(user.Room, bson.ObjectIdHex(roomID))}}
// 		userID := user.UserID
// 		err = roomMongo.ConnectionDB.DB(dbName).C(collectionUser).UpdateId(userID, newUser)
// 		if err != nil {
// 			err = fmt.Errorf("error updating user %s", err)
// 			break
// 		}
// 	}
// 	return err
// }

// func (roomMongo RoomRepositoryMongo) DeleteMemberFromRoom(roomID string, userID []string) error {
// 	// TODO might need to fix logic
// 	var ConnectionDB, err = mgo.Dial(utills.MONGOENDPOINT)
// 	if err != nil {
// 		return err
// 	}
// 	// for delete in room
// 	var room model.Room
// 	err = ConnectionDB.DB(dbName).C(collectionRoom).FindId(bson.ObjectId(roomID)).One(&room)

// 	// TODO fix this, i just want it to compile for now
// 	NewListString := utills.RemoveFormListBson(room.ListUser, utills.ToObjectIdArr(userID)[0])
// 	newUser := bson.M{"$set": bson.M{"listUser": NewListString}}
// 	ConnectionDB.DB(dbName).C(collectionRoom).UpdateId(roomID, newUser)
// 	// for delete in user
// 	var user model.User
// 	err = ConnectionDB.DB(dbName).C(collectionRoom).FindId(utills.ToObjectIdArr(userID)).One(&user)
// 	NewListString = utills.RemoveFormListBson(user.Room, bson.ObjectIdHex(roomID))
// 	newUser = bson.M{"$set": bson.M{"room": NewListString}}
// 	ConnectionDB.DB("User").C("UserData").UpdateId(userID, newUser)
// 	return nil
// }
