package main

import (
	"context"
	"flag"
	"log"
	"net"
	"proxySenior/share/proto"
	"time"

	"github.com/globalsign/mgo/bson"

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
	MessageID bson.ObjectId `bson:"_id"`
	TimeStamp time.Time     `bson:"timestamp"`
	RoomID    bson.ObjectId `bson:"roomId"`
	UserID    bson.ObjectId `bson:"userId"`
	ClientUID string        `bson:"clientUID"`
	Data      string        `bson:"data"`
	Type      string        `bson:"type"`
}

func (b *BackupServer) OnMessageIn(context context.Context, chat *proto.Chat) (*proto.Empty, error) {
	log.Println("Access OnMessageIn")
	bMsg := BackupMessage{
		MessageID: bson.ObjectIdHex(chat.MessageId),
		TimeStamp: time.Unix(chat.Timestamp, 0),
		RoomID:    bson.ObjectIdHex(chat.RoomId),
		UserID:    bson.ObjectIdHex(chat.UserId),
		ClientUID: chat.ClientUid,
		Data:      chat.Data,
		Type:      chat.Type,
	}
	log.Println("Incoming Message  >>>>>>", bMsg, "\n")
	var message []BackupMessage
	conn.DB("backup").C("message").Find(nil).All(&message)
	for _, v := range message {
		log.Println(v, "\n")
	}
	return &proto.Empty{}, nil
	// return &proto.Empty{}, conn.DB("backup").C("message").Insert(bMsg)

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
	log.Println("Start-go Server with PORT", "5005")
	go func() {
		var err error
		conn, err = mgo.Dial("mongodb://localhost:27017")
		if err != nil {
			log.Fatal("error running", err)
		}
		isReady = true
	}()

	flag.Parse()
	lis, err := net.Listen("tcp", "localhost:5005")
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
