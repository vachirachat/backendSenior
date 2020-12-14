package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type MessageRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

type messageMongoDB struct {
	MessageID bson.ObjectId `json:"messageId" bson:"_id,omitempty"`
	TimeStamp time.Time     `json:"timestamp" bson:"timestamp"`
	RoomID    bson.ObjectId `json:"roomId" bson:"roomId"`
	UserID    bson.ObjectId `json:"userId" bson:"userId"`
	Data      string        `json:"data" bson:"data"`
	Type      string        `json:"type" bson:"type"`
}

func toCreateMessageMongoDB(msg model.Message) messageMongoDB {
	var msgDB messageMongoDB
	msgDB.MessageID = bson.NewObjectId()
	msgDB.TimeStamp = msg.TimeStamp
	msgDB.RoomID = bson.ObjectIdHex(msg.RoomID)
	msgDB.UserID = bson.ObjectIdHex(msg.RoomID)
	msgDB.Data = msg.Data
	msgDB.Type = msg.Type
	return msgDB
}

func toCreateMessageMongoDBArr(messages []model.Message) []messageMongoDB {
	var msgDBs = make([]messageMongoDB, len(messages))
	for i := range messages {
		msgDBs[i] = toCreateMessageMongoDB(messages[i])
	}
	return msgDBs
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
	messages = model.MessageListToMongoID(messages)
	return messages, err
}

// GetMessagesByRoom return messages from specified room, with optional time filter
func (messageMongo *MessageRepositoryMongo) GetMessagesByRoom(roomID string, timeRange *model.TimeRange) ([]model.Message, error) {
	var messages []model.Message
	// filter := queryFromTimeRange(timeRange)
	filter := bson.M{}
	filter["roomId"] = bson.ObjectIdHex(roomID)
	err := messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).Find(filter).All(&messages)
	messages = model.MessageListToMongoID(messages)
	return messages, err
}

// GetMessageByID return message by id
func (messageMongo *MessageRepositoryMongo) GetMessageByID(messageID string) (model.Message, error) {
	var message model.Message
	err := messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).FindId(bson.ObjectIdHex(messageID)).One(&message)
	return message.MessageStringIDToMongoID(), err
}

// AddMessage insert message
func (messageMongo *MessageRepositoryMongo) AddMessage(message model.Message) (string, error) {
	cnt, err := messageMongo.ConnectionDB.DB(dbName).C(collectionRoom).FindId(bson.ObjectIdHex(message.RoomID)).Count()
	if cnt == 0 || err != nil {
		return "", fmt.Errorf("room error: %s", err)
	}
	err = messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).Insert(message)
	return bson.ObjectIdHex(message.RoomID).Hex(), err
}

// DeleteMessageByID delete message by id
func (messageMongo *MessageRepositoryMongo) DeleteMessageByID(messageID string) error {
	return messageMongo.ConnectionDB.DB(dbName).C(collectionMessage).RemoveId(bson.ObjectIdHex(messageID))
}
