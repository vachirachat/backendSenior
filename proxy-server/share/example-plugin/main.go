package main

import (
	"log"
	"proxySenior/share/backup"
	"proxySenior/share/config"

	"github.com/globalsign/mgo"
	"github.com/hashicorp/go-plugin"
)

var conn *mgo.Session
var isReady bool

type Backup struct{}

var _ backup.BackupService = (*Backup)(nil)

type BackupMessage struct {
	MessageID string `bson:"_id"`
	TimeStamp int64  `bson:"timestamp"`
	RoomID    string `bson:"roomId"`
	UserID    string `bson:"userId"`
	ClientUID string `bson:"clientUID"`
	Data      string `bson:"data"`
	Type      string `bson:"type"`
}

func (b *Backup) OnMessageIn(msg backup.RawMessage) error {
	bMsg := BackupMessage{
		MessageID: msg.MessageID,
		TimeStamp: msg.TimeStamp,
		RoomID:    msg.RoomID,
		UserID:    msg.UserID,
		ClientUID: msg.ClientUID,
		Data:      msg.Data,
		Type:      msg.Type,
	}
	return conn.DB("backup").C("message").Insert(bMsg)
}

func (b *Backup) IsReady() (bool, error) {
	return isReady, nil
}

func main() {
	// change config here
	go func() {
		var err error
		conn, err = mgo.Dial("mongodb://localhost:27017")
		if err != nil {
			log.Fatal("error running", err)
		}
		isReady = true
	}()

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: config.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"backup": &backup.BackupGRPCPlugin{Impl: &Backup{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
