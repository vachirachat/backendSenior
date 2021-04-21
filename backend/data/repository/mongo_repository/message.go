package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type MessageRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

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
func (messageMongo *MessageRepositoryMongo) GetAllMessages(timeRange *model.TimeRange) ([]model.Message, error) {
	var messages []model.Message
	err := messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).Find(queryFromTimeRange(timeRange)).Limit(1000).All(&messages)
	return messages, err
}

// GetMessagesByRoom return messages from specified room, with optional time filter
func (messageMongo *MessageRepositoryMongo) GetMessagesByRoom(roomID string, timeRange *model.TimeRange) ([]model.Message, error) {
	var messages []model.Message
	filter := queryFromTimeRange(timeRange)
	filter["roomId"] = bson.ObjectIdHex(roomID)
	err := messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).Find(filter).Limit(1000).Sort("-timestamp").All(&messages)
	return messages, err
}

// GetMessageByID return message by id
func (messageMongo *MessageRepositoryMongo) GetMessageByID(messageID string) (model.Message, error) {
	var message model.Message
	err := messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).FindId(bson.ObjectIdHex(messageID)).One(&message)
	return message, err
}

// AddMessage insert message
func (messageMongo *MessageRepositoryMongo) AddMessage(message model.Message) (string, error) {
	message.MessageID = bson.NewObjectId()
	cnt, err := messageMongo.ConnectionDB.DB(dbName).C(collectionRoom).FindId(message.RoomID).Count()
	if cnt == 0 || err != nil {
		return "", fmt.Errorf("room error: %s", err)
	}
	err = messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).Insert(message)
	return message.MessageID.Hex(), err
}

// DeleteMessageByID delete message by id
func (messageMongo *MessageRepositoryMongo) DeleteMessageByID(messageID string) error {
	objectID := bson.ObjectIdHex(messageID)
	return messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).RemoveId(objectID)
}
