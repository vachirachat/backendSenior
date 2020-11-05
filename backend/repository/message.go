package repository

import (
	"backendSenior/model"
	"backendSenior/utills"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type MessageRepository interface {
	GetAllMessage() ([]model.Message, error)
	// GetLastMessage() (model.Message, error)
	GetMessageByID(userID string) (model.Message, error)
	AddMessage(message model.Message) error
	DeleteMessageByID(userID string) error
}

type MessageRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

type subscription struct {
	clientName string
	clientID   string
	room       string
}

const (
	DBMessage         = "Message"
	collectionMessage = "MessageData"
)

func (messageMongo MessageRepositoryMongo) GetAllMessage() ([]model.Message, error) {
	var Messages []model.Message
	err := messageMongo.ConnectionDB.DB(DBMessage).C(collectionMessage).Find(nil).All(&Messages)
	return Messages, err
}

func (messageMongo MessageRepositoryMongo) GetMessageByID(messageID string) (model.Message, error) {
	var message model.Message
	objectID := bson.ObjectIdHex(messageID)
	err := messageMongo.ConnectionDB.DB(DBMessage).C(collectionMessage).Find(objectID).One(&message)
	return message, err
}

func (messageMongo MessageRepositoryMongo) AddMessage(message model.Message) error {
	return messageMongo.ConnectionDB.DB(DBMessage).C(collectionMessage).Insert(message)
}

func (messageMongo MessageRepositoryMongo) DeleteMessageByID(messageID string) error {
	objectID := bson.ObjectIdHex(messageID)
	return messageMongo.ConnectionDB.DB(DBMessage).C(collectionMessage).RemoveId(objectID)
}

func AddMessageDB(message []byte, room string, clientID string, clientName string) error {
	var ConnectionDB, _ = mgo.Dial(utills.MONGOENDPOINT)
	var messageDB model.Message
	messageDB.TimeStamp = time.Now()
	messageDB.RoomID = room
	messageDB.UserID = clientID
	messageDB.Name = clientName
	messageDB.Data = string(message)
	messageDB.Type = "Send"
	return ConnectionDB.DB(DBMessage).C(collectionMessage).Insert(messageDB)
}
