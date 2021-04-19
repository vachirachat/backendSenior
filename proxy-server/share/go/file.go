package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"go-module/proto"
	"log"
	"net"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/mergermarket/go-pkcs7"

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
	for i, v := range message {
		if i > len(message)-5 {
			log.Println(v, "\n")
		}

	}
	// return &proto.Empty{}, nil
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
		conn, err = mgo.Dial("172.17.0.2:27017")
		if err != nil {
			log.Fatal("error running", err)
		}
		isReady = true
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf("172.17.0.4:%d", 5005))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	opts = []grpc.ServerOption{}
	log.Println("Start-go Server with PORT", "5005")
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterBackupServer(grpcServer, NewBackupServer())
	grpcServer.Serve(lis)
}

var Keymap = "abcdefghijklmnopqrstuvwxyz012345"

// Decrypt takes a message, then return message with data decrypted with appropiate key
func (b *BackupServer) EncryptedMessage(ctx context.Context, chat *proto.Chat) (*proto.Chat, error) {
	key := []byte(Keymap)

	// b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(message.Data)))
	cipherText, err := base64.StdEncoding.DecodeString(chat.Data)
	if err != nil {
		return &proto.Chat{}, fmt.Errorf("decode b64: %s", err.Error())
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return &proto.Chat{}, err
	}

	if len(cipherText) < aes.BlockSize {
		return &proto.Chat{}, errors.New("cipher text too short")
	}

	iv := cipherText[:aes.BlockSize]
	data := cipherText[aes.BlockSize:]
	if len(data)%aes.BlockSize != 0 {
		return &proto.Chat{}, errors.New("wrong cipher text size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	if err != nil {
		return &proto.Chat{}, err
	}

	decrypted := make([]byte, len(data))

	mode.CryptBlocks(decrypted, data)

	decrypted, _ = pkcs7.Unpad(decrypted, aes.BlockSize)
	chat.Data = string(decrypted)

	return chat, nil
}
