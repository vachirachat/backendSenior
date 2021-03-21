package main

/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a simple gRPC client that demonstrates how to use gRPC-Go libraries
// to perform unary, client streaming, server streaming and full duplex RPCs.
//
// It interacts with the route guide service whose definition can be found in routeguide/route_guide.proto.

import (
	context "context"
	"fmt"
	"log"
	"proxySenior/share/proto"
	"strconv"
	"time"

	"github.com/globalsign/mgo/bson"
	"google.golang.org/grpc"
)

// var serverAddr = flag.String("server_addr", "localhost:5005", "The server address in the format of host:port")

// RawMessage is message received over GRPC
type RawMessage struct {
	MessageID string `bson:"_id"`
	TimeStamp int64  `bson:"timestamp"`
	RoomID    string `bson:"roomId"`
	UserID    string `bson:"userId"`
	ClientUID string `bson:"clientUID"`
	Data      string `bson:"data"`
	Type      string `bson:"type"`
}
type Message struct {
	MessageID bson.ObjectId `json:"messageId" bson:"_id,omitempty"`
	TimeStamp time.Time     `json:"timestamp" bson:"timestamp"`
	RoomID    bson.ObjectId `json:"roomId" bson:"roomId"`
	UserID    bson.ObjectId `json:"userId" bson:"userId"`
	ClientUID string        `json:"uid" bson:"uid"`
	Data      string        `json:"data" bson:"data"`
	Type      string        `json:"type" bson:"type"`
}

// onMessage
func onMessage(client proto.BackupClient, msg RawMessage) {
	log.Println("Getting onMessage for point", msg.Data)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	log.Println(ctx)
	defer cancel()
	feature, err := client.OnMessageIn(ctx, &proto.Chat{
		MessageId: msg.MessageID,
		RoomId:    msg.RoomID,
		UserId:    msg.UserID,
		Timestamp: msg.TimeStamp,
		Type:      msg.Type,
		ClientUid: msg.ClientUID,
		Data:      msg.Data,
	})
	if err != nil {
		log.Fatalf("%v.onMessage(_) = _, %v: ", client, err)
	}
	log.Println(feature)
}

func isReady(client proto.BackupClient) *proto.Status {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ok, err := client.IsReady(ctx, &proto.Empty{})
	if err != nil {
		log.Fatalf("%v.isReady(_) = _, %v: ", client, err)
	}

	log.Println(ok)
	return ok
}

func Encryption(client proto.BackupClient, msg RawMessage) {
	log.Println("Getting Encryption for point", msg.Data)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	feature, err := client.EncryptedMessage(ctx, &proto.Chat{
		MessageId: msg.MessageID,
		RoomId:    msg.RoomID,
		UserId:    msg.UserID,
		Timestamp: msg.TimeStamp,
		Type:      msg.Type,
		ClientUid: msg.ClientUID,
		Data:      msg.Data,
	})
	if err != nil {
		log.Fatalf("%v.Encryption(_) = _, %v: ", client, err)
	}
	log.Println("\n EncryptedMessage >>", feature)

	feature, err = client.DecryptedMessage(ctx, feature)
	if err != nil {
		log.Fatalf("%v.Decrypted(_) = _, %v: ", client, err)
	}

	log.Println(feature.Timestamp)

	i, err := strconv.ParseInt(fmt.Sprint(feature.Timestamp), 10, 64)
	if err != nil {
		panic(err)
	}

	tm := time.Unix(i, 0)

	msgDe := Message{
		MessageID: bson.ObjectIdHex(feature.MessageId),
		TimeStamp: tm,
		RoomID:    bson.ObjectIdHex("60001d1cf0a50a974cee376d"),
		UserID:    bson.ObjectIdHex("60001e33584cb6da2059f5b7"),
		ClientUID: "60001d1cf0a50a974cee376d",
		// Data:      "Test GO Plugin Server",
		Data: ">>> Hello docker Server",
		// Data: "Test Python Plugin Server",
		// Data: "Test JS Plugin Server",
		Type: "CHAT",
	}

	log.Println("\n DecryptedMessage >>", msgDe)

}

func getKey(client proto.BackupClient) {
	log.Println("Getting Encryption for point getKey")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	feature, err := client.GetKey(ctx, &proto.Empty{})
	if err != nil {
		log.Fatalf("%v.getKey(_) = _, %v: ", client, err)
	}
	log.Println("\n GetKey >>", feature)
}

func main() {
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	// log.Println("Connect Port localhost:7000")
	// conn, err := grpc.Dial(":6000", opts...)

	log.Println("Connect Port localhost:3333")
	conn, err := grpc.Dial("localhost:3333", opts...)

	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	client := proto.NewBackupClient(conn)
	log.Println(client)

	msg := RawMessage{
		MessageID: bson.NewObjectId().Hex(),
		TimeStamp: time.Now().Unix(),
		RoomID:    bson.ObjectIdHex("60001d1cf0a50a974cee376d").Hex(),
		UserID:    bson.ObjectIdHex("60001e33584cb6da2059f5b7").Hex(),
		ClientUID: "60001d1cf0a50a974cee376d",
		// Data:      "Test GO Plugin Server",
		Data: ">>> Hello docker Server",
		// Data: "Test Python Plugin Server",
		// Data: "Test JS Plugin Server",
		Type: "CHAT",
	}

	defer conn.Close()

	ok := isReady(client)
	log.Print("Return from isReady", ok)
	// onMessage(client, msg)
	Encryption(client, msg)
	getKey(client)
}

// docker run --rm -itd -p 27017:27017 -v /Users/waritphon/Downloads/Senior:/data/db mongo

// chat-route-handler
// userID := "60001d1cf0a50a974cee376d"
// routerGroup.GET("/ws" /*, handler.authMiddleware.AuthRequired() */, func(context *gin.Context) {

// // Prin
// var message []BackupMessage
// conn.DB("backup").C("message").Find(nil).All(&message)
// for _, v := range message {
// 	log.Println(v, "\n")
// }
// return &proto.Empty{}, nil

// IN main client-> proxy
// // TO TEST must DELETE : TEST Message
// Message := backup.RawMessage{
// 	MessageID: bson.ObjectId("60001d1cf0a50a974cee376d").Hex(),
// 	TimeStamp: time.Now().Unix(),
// 	RoomID:    bson.ObjectIdHex("60001d1cf0a50a974cee376d").Hex(),
// 	UserID:    bson.ObjectIdHex("60001e33584cb6da2059f5b7").Hex(),
// 	ClientUID: "60001d1cf0a50a974cee376d",
// 	Data:      "Test-message-1",
// 	Type:      "CHAT",
// }

// User in OnMessageIn on_message_plugin_port
// feature, err := temp.OnMessageIn(ctx, &proto.Chat{
// 	MessageId: msg.MessageID,
// 	RoomId:    msg.RoomID,
// 	Timestamp: msg.TimeStamp,
// 	UserId:    msg.UserID,
// 	Type:      msg.Type,
// 	ClientUid: msg.ClientUID,
// 	Data:      msg.Data,
// })

// -> simple websocket-check
// ws://localhost:8090/api/v1/chat/ws?userId=60001d1cf0a50a974cee376d
// {"type":"CHAT",
// "payload":{
// "data":"hi 1149 --> Test in DB -> TO javascript new",
// "uid":"60007606000000000400a304",
// "roomId":"60001e33584cb6da2059f5b7",
// "userId":"60001d1cf0a50a974cee376d",
// "type":"TEXT"}
// }
