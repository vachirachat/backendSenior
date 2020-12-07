package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type MessageRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

const (
	DBMessage         = "Message"
	collectionMessage = "MessageData"
)

var _ repository.MessageRepository = (*MessageRepositoryMongo)(nil)

// GetAllMessages return all message from all rooms
func (messageMongo MessageRepositoryMongo) GetAllMessages() ([]model.Message, error) {
	var Messages []model.Message
	err := messageMongo.ConnectionDB.DB(DBMessage).C(collectionMessage).Find(nil).All(&Messages)
	return Messages, err
}

// GetMessageByID return message by id
func (messageMongo MessageRepositoryMongo) GetMessageByID(messageID string) (model.Message, error) {
	var message model.Message
	objectID := bson.ObjectIdHex(messageID)
	err := messageMongo.ConnectionDB.DB(DBMessage).C(collectionMessage).Find(objectID).One(&message)
	return message, err
}

// AddMessage insert message
func (messageMongo MessageRepositoryMongo) AddMessage(message model.Message) error {
	return messageMongo.ConnectionDB.DB(DBMessage).C(collectionMessage).Insert(message)
}

// DeleteMessageByID delete message by id
func (messageMongo MessageRepositoryMongo) DeleteMessageByID(messageID string) error {
	objectID := bson.ObjectIdHex(messageID)
	return messageMongo.ConnectionDB.DB(DBMessage).C(collectionMessage).RemoveId(objectID)
}
