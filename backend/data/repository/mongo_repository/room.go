package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/utills"
	"fmt"
	"log"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

type RoomRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

type roomMongoDB struct {
	RoomID           bson.ObjectId   `json:"roomId" bson:"_id,omitempty"`
	RoomName         string          `json:"roomName" bson:"roomName"`
	CreatedTimeStamp time.Time       `json:"-" bson:"createdTimestamp"`
	RoomType         string          `json:"roomType" bson:"roomType"`
	ListUser         []bson.ObjectId `json:"listUser" bson:"listUser"`
}

var _ repository.RoomRepository = (*RoomRepositoryMongo)(nil)

func toCreateRoomMongoDB(room model.Room) roomMongoDB {
	var rooomDB roomMongoDB
	rooomDB.RoomID = bson.NewObjectId()
	rooomDB.RoomName = room.RoomName
	rooomDB.CreatedTimeStamp = room.CreatedTimeStamp
	rooomDB.RoomType = room.RoomType
	rooomDB.ListUser = utills.ToObjectIdArr(room.ListUser)
	return rooomDB
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
	return toCreateRoomMongoDB(room).RoomID.String(), roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).Insert(toCreateRoomMongoDB(room))
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
	fmt.Println(roomID)
	bid := bson.ObjectIdHex(roomID)
	fmt.Println(bid)
	err := roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).UpdateId(bid, bson.M{
		"$push": bson.M{
			"listUser": bson.M{
				"$each": utills.ToObjectIdArr(listUser), // add all from listUser to array
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error updating room users %s", err)
	}

	if err != nil {
		return fmt.Errorf("error re-getting room users %s", err)
	}

	// var room model.Room
	// log.Println(bson.ObjectId(roomID))
	// log.Println(bson.ObjectIdHex(roomID))
	// err = roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).FindId(bson.ObjectId(roomID)).One(&room)
	// if err != nil {
	// 	log.Println("line 101 Error :: err = roomMongo.ConnectionDB.DB(dbName).C(collectionRoom).FindId")
	// 	return err
	// }
	for _, s := range listUser {
		var user model.User
		err = roomMongo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.ObjectIdHex(s)).One(&user)
		didntAdd := false
		for _, v := range user.Room {
			if bson.ObjectIdHex(v) == bson.ObjectIdHex(roomID) {
				didntAdd = true
			}
		}
		if didntAdd == false {
			newUser := bson.M{"$set": bson.M{"room": append(utills.ToObjectIdArr(user.Room), bson.ObjectIdHex(roomID))}}
			userID := bson.ObjectIdHex(s)
			err = roomMongo.ConnectionDB.DB(dbName).C(collectionUser).UpdateId(userID, newUser)
		}

	}

	return err
}

func (roomMongo RoomRepositoryMongo) DeleteMemberFromRoom(roomID string, userIDs []string) error {
	// TODO might need to fix logic
	var ConnectionDB, err = mgo.Dial(utills.MONGOENDPOINT)
	if err != nil {
		return err
	}
	// for delete in room
	var room model.Room
	err = ConnectionDB.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(roomID)).One(&room)
	room = room.RoomStringIDToMongoID()
	log.Println(room)
	var NewListString []bson.ObjectId
	// TODO fix this, i just want log.Println for fix error didnt use
	for _, v := range utills.ToObjectIdArr(userIDs) {
		NewListString := utills.RemoveFormListBson(utills.ToObjectIdArr(room.ListUser), v)
		log.Println(NewListString)
	}
	newUser := bson.M{"$set": bson.M{"listUser": NewListString}}
	ConnectionDB.DB(dbName).C(collectionRoom).UpdateId(bson.ObjectIdHex(roomID), newUser)
	// for delete in user
	// var user model.User
	// err = ConnectionDB.DB(collectionUser).C(collectionUser).FindId(userID).One(&user)
	// roomIDString := roomID
	log.Println(utills.ToObjectIdArr(userIDs))
	for _, v := range utills.ToObjectIdArr(userIDs) {
		var user model.User
		err = ConnectionDB.DB(dbName).C(collectionUser).FindId(v).One(&user)
		user = user.UserStringIDToMongoID()
		NewListString = utills.RemoveFormListBson(utills.ToObjectIdArr(user.Room), bson.ObjectIdHex(roomID))
		newUser = bson.M{"$set": bson.M{"room": NewListString}}
		ConnectionDB.DB(dbName).C(collectionUser).UpdateId(v, newUser)
	}

	return nil
}
