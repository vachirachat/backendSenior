package external

import (
	"backendSenior/domain/model"
	"context"
	"fmt"
	"log"
	"proxySenior/share/proto"
	"time"

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

	conn, err := grpc.Dial("localhost:5005", opts...)
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	for {
		temp := *obp.client
		ok, err := temp.IsReady(ctx, &proto.Empty{})
		if err != nil {
			log.Println(">>>>> ", err)
			return err
		}
		if ok.GetOk() {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	defer cancel()
	return nil
}

// GetService return instance of backup service to be called
// func (p *GRPCOnPortMessagePlugin) GetService() backup.BackupService {
// 	return p.client
// }

// OnMessageIn convert message from model.Message then send over GRPC
// TO TEST must DELETE : Change backup to Model.Message
func (obp *GRPCOnPortMessagePlugin) OnMessageIn(msg model.Message) error {
	log.Println("Getting onMessage for point", msg.Data)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	temp := *obp.client
	feature, err := temp.OnMessageIn(ctx, &proto.Chat{
		MessageId: msg.MessageID.Hex(),
		RoomId:    msg.RoomID.Hex(),
		Timestamp: msg.TimeStamp.Unix(),
		UserId:    msg.UserID.Hex(),
		Type:      msg.Type,
		ClientUid: msg.ClientUID,
		Data:      msg.Data,
	})

	if err != nil {
		log.Fatalf("%v.onMessage(_) = _, %v: ", obp.client, err)
	}
	log.Println(feature)
	defer cancel()
	return err
}
