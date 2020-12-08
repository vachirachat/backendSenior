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
	dbMessage         = "mychat"
	collectionMessage = "messages"
)

var _ repository.MessageRepository = (*MessageRepositoryMongo)(nil)

func queryFromTimeRange(rng *model.TimeRange) map[string]interface{} {
	filter := bson.M{}
	if rng == nil {
		return filter
	}
	if !rng.To.IsZero() {
		filter["$lte"] = rng.To
	}
	if !rng.From.IsZero() {
		filter["$gte"] = rng.From
	}
	return filter
}

// GetAllMessages return all message from all rooms with optional time filter
func (messageMongo MessageRepositoryMongo) GetAllMessages(timeRange *model.TimeRange) ([]model.Message, error) {
	var messages []model.Message
	err := messageMongo.ConnectionDB.DB(dbMessage).C(collectionMessage).Find(queryFromTimeRange(timeRange)).All(&messages)
	return messages, err
}

// GetMessagesByRoom return messages from specified room, with optional time filter
func (messageMongo *MessageRepositoryMongo) GetMessagesByRoom(roomID string, timeRange *model.TimeRange) ([]model.Message, error) {
	var messages []model.Message
	filter := queryFromTimeRange(timeRange)
	filter["roomId"] = roomID
	err := messageMongo.ConnectionDB.DB(dbMessage).C(collectionMessage).Find(filter).All(&messages)
	return messages, err
}

// GetMessageByID return message by id
func (messageMongo MessageRepositoryMongo) GetMessageByID(messageID string) (model.Message, error) {
	var message model.Message
	err := messageMongo.ConnectionDB.DB(dbMessage).C(collectionMessage).FindId(messageID).One(&message)
	return message, err
}

// AddMessage insert message
func (messageMongo MessageRepositoryMongo) AddMessage(message model.Message) (string, error) {
	message.MessageID = bson.NewObjectId()
	err := messageMongo.ConnectionDB.DB(dbMessage).C(collectionMessage).Insert(message)
	return message.MessageID.Hex(), err
}

// DeleteMessageByID delete message by id
func (messageMongo MessageRepositoryMongo) DeleteMessageByID(messageID string) error {
	objectID := bson.ObjectIdHex(messageID)
	return messageMongo.ConnectionDB.DB(dbMessage).C(collectionMessage).RemoveId(objectID)
}
