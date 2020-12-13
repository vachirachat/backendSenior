package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type MessageRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

var _ repository.MessageRepository = (*MessageRepositoryMongo)(nil)

func queryFromTimeRange(timeRange *model.TimeRange) map[string]interface{} {
	filter := bson.M{}
	tempFilter := bson.M{}
	// comment: case nil &timeRange
	if timeRange == nil {
		return filter
	}
	timeRangeValue := *timeRange
	// TODO :: TEST GET MESSAGE
	// rangeTimeFrom, _ := time.Parse("UnixDate", timeRangeValue.To.String())
	// rangeTimeTo, _ := time.Parse("UnixDate", timeRangeValue.From.String())
	if !timeRangeValue.To.IsZero() {
		tempFilter["$lte"] = timeRangeValue.To
	}
	if !timeRangeValue.From.IsZero() {
		tempFilter["$gte"] = timeRangeValue.From
	}
	filter["timestamp"] = tempFilter
	return filter
}

// GetAllMessages return all message from all rooms with optional time filter
func (messageMongo *MessageRepositoryMongo) GetAllMessages(timeRange *model.TimeRange) ([]model.Message, error) {
	var messages []model.Message
	err := messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).Find(queryFromTimeRange(timeRange)).All(&messages)
	return messages, err
}

// GetMessagesByRoom return messages from specified room, with optional time filter
func (messageMongo *MessageRepositoryMongo) GetMessagesByRoom(roomID string, timeRange *model.TimeRange) ([]model.Message, error) {
	var messages []model.Message
	filter := queryFromTimeRange(timeRange)
	filter["roomId"] = bson.ObjectIdHex(roomID)
	err := messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).Find(filter).All(&messages)
	log.Println(messages)
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
	err := messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).Insert(message)
	return message.MessageID.Hex(), err
}

// DeleteMessageByID delete message by id
func (messageMongo *MessageRepositoryMongo) DeleteMessageByID(messageID string) error {
	return messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).RemoveId(bson.ObjectIdHex(messageID))
}
