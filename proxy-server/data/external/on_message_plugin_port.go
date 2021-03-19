package external

import (
	"backendSenior/domain/model"
	"context"
	"fmt"
	"log"
	"proxySenior/share/proto"
	"strconv"
	"time"

	"github.com/globalsign/mgo/bson"
	"google.golang.org/grpc"
)

type GRPCOnPortMessagePlugin struct {
	client *proto.BackupClient
	conn   *grpc.ClientConn
}

func NewGRPCOnPortMessagePlugin(serverPort string) *GRPCOnPortMessagePlugin {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(serverPort, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	client := proto.NewBackupClient(conn)

	return &GRPCOnPortMessagePlugin{client: &client}
}

func (obp *GRPCOnPortMessagePlugin) CloseConnection() {
	obp.conn.Close()
	return
}

// Wait blocks until underlying GRPC server is ready
func (obp *GRPCOnPortMessagePlugin) Wait() error {
	fmt.Println("waiting for GRPC server...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	for {
		temp := *obp.client
		ok, err := temp.IsReady(ctx, &proto.Empty{})
		log.Println("Return temp", ok)
		if err != nil {
			return err
		}
		if ok.GetOk() {
			return nil
		}
		time.Sleep(5 * time.Second)
	}

}

// GetService return instance of backup service to be called
// func (p *GRPCOnPortMessagePlugin) GetService() backup.BackupService {
// 	return p.client
// }

// OnMessageIn convert message from model.Message then send over GRPC
func (obp *GRPCOnPortMessagePlugin) OnMessageIn(msg model.Message) error {
	log.Println("Getting onMessage for point", msg.Data)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// Create sending Messasge Format
	client := *obp.client
	_, err := client.OnMessageIn(ctx, &proto.Chat{
		MessageId: msg.MessageID.Hex(),
		RoomId:    msg.RoomID.Hex(),
		Timestamp: msg.TimeStamp.Unix(),
		UserId:    msg.UserID.Hex(),
		Type:      msg.Type,
		ClientUid: msg.ClientUID,
		Data:      msg.Data,
	})

	if err != nil {
		log.Fatalf("%v.onMessageIn %v: ", obp.client, err)
	}
	defer cancel()
	return err
}

// CustomEncryption convert message from model.Message then send over GRPC >> to Manage encryption
func (obp *GRPCOnPortMessagePlugin) CustomEncryption(msg model.Message) (model.Message, error) {
	log.Println("Getting ForwardEncrypt for point", msg.Data)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := *obp.client
	message, err := client.EncryptedMessage(ctx, &proto.Chat{
		MessageId: msg.MessageID.Hex(),
		RoomId:    msg.RoomID.Hex(),
		UserId:    msg.UserID.Hex(),
		// Timestamp: msg.TimeStamp.UnixNano() / int64(time.Millisecond),
		Timestamp: msg.TimeStamp.Unix(),
		Type:      msg.Type,
		ClientUid: msg.ClientUID,
		Data:      msg.Data,
	})

	if err != nil {
		log.Fatalf("%v.Encryption(_) = _, %v: ", client, err)
	}
	log.Println("\n EncryptedMessage >>", message)

	// REFACTOR : Temp Chat.timestamp -> Message.Timestamp
	i, err := strconv.ParseInt(fmt.Sprint(message.Timestamp), 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)
	return model.Message{
		MessageID: bson.ObjectIdHex(message.MessageId),
		TimeStamp: tm,
		RoomID:    bson.ObjectIdHex(message.RoomId),
		UserID:    bson.ObjectIdHex(message.UserId),
		ClientUID: message.ClientUid,
		Data:      message.Data,
		Type:      message.Type,
	}, nil
}

func (obp *GRPCOnPortMessagePlugin) CustomDecryption(msg model.Message) (model.Message, error) {
	log.Println("Getting ForwardDecrypt for point", msg.Data)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := *obp.client
	// REFACTOR : Temp Chat.timestamp -> Message.Timestamp

	message, err := client.DecryptedMessage(ctx, &proto.Chat{
		MessageId: msg.MessageID.Hex(),
		RoomId:    msg.RoomID.Hex(),
		UserId:    msg.UserID.Hex(),
		// Timestamp: msg.TimeStamp.UnixNano() / int64(time.Millisecond),
		Timestamp: msg.TimeStamp.Unix(),
		Type:      msg.Type,
		ClientUid: msg.ClientUID,
		Data:      msg.Data,
	})
	if err != nil {
		log.Fatalf("%v.Encryption(_) = _, %v: ", client, err)
	}

	i, err := strconv.ParseInt(fmt.Sprint(message.Timestamp), 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)

	log.Println("\n DecryptedMessage >>", message)
	return model.Message{
		MessageID: bson.ObjectIdHex(message.MessageId),
		TimeStamp: tm,
		RoomID:    bson.ObjectIdHex(message.RoomId),
		UserID:    bson.ObjectIdHex(message.UserId),
		ClientUID: message.ClientUid,
		Data:      message.Data,
		Type:      message.Type,
	}, nil
}
