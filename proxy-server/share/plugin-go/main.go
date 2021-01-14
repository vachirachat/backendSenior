package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"proxySenior/share/proto"

	"github.com/globalsign/mgo"
	"google.golang.org/grpc"
)

var isReady bool
var conn *mgo.Session

type BackupServer struct {
	proto.UnimplementedBackupServer
	saveChats []*proto.Chat // read-only Chat after initialized
}

type BackupMessage struct {
	MessageID string `bson:"_id"`
	TimeStamp int64  `bson:"timestamp"`
	RoomID    string `bson:"roomId"`
	UserID    string `bson:"userId"`
	ClientUID string `bson:"clientUID"`
	Data      string `bson:"data"`
	Type      string `bson:"type"`
}

func (b *BackupServer) OnMessageIn(context context.Context, chat *proto.Chat) (*proto.Empty, error) {
	log.Println("Access OnMessage")
	log.Println("Access ", chat)
	bMsg := BackupMessage{
		MessageID: chat.MessageId,
		TimeStamp: chat.Timestamp,
		RoomID:    chat.RoomId,
		UserID:    chat.UserId,
		ClientUID: chat.ClientUid,
		Data:      chat.Data,
		Type:      chat.Type,
	}
	return &proto.Empty{}, conn.DB("backup").C("message").Insert(bMsg)
}
func (b *BackupServer) IsReady(context context.Context, empty *proto.Empty) (*proto.Status, error) {
	log.Println("Access IsReady")
	log.Println("Access ", empty)
	return &proto.Status{Ok: true}, nil
}

func NewBackupServer() proto.BackupServer {
	return &BackupServer{saveChats: make([]*proto.Chat, 0)}
}

func main() {
	//connect mongo Server
	go func() {
		var err error
		conn, err = mgo.Dial("mongodb://localhost:27017")
		if err != nil {
			log.Fatal("error running", err)
		}
		isReady = true
	}()

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 5005))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	opts = []grpc.ServerOption{}

	grpcServer := grpc.NewServer(opts...)
	proto.RegisterBackupServer(grpcServer, NewBackupServer())
	grpcServer.Serve(lis)
}

// var conn *mgo.Session
// var isReady bool

// type Backup struct{}

// var _ backup.BackupService = (*Backup)(nil)

// func (b *Backup) OnMessageIn(msg backup.RawMessage) error {

// }

// func (b *Backup) IsReady() (bool, error) {
// 	return isReady, nil
// }

// func main() {
// 	// change config here
// go func() {
// 	var err error
// 	conn, err = mgo.Dial("mongodb://localhost:27017")
// 	if err != nil {
// 		log.Fatal("error running", err)
// 	}
// 	isReady = true
// }()

// 	proto.Serve(&proto.ServeConfig{
// 		HandshakeConfig: config.HandshakeConfig,
// 		protos: map[string]proto.proto{
// 			"backup": &backup.BackupGRPCproto{Impl: &Backup{}},
// 		},
// 		GRPCServer: proto.DefaultGRPCServer,
// 	})
// }
