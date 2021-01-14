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
	"flag"
	"log"
	"proxySenior/share/proto"
	"time"

	"google.golang.org/grpc"
)

var serverAddr = flag.String("server_addr", "localhost:5005", "The server address in the format of host:port")

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

// onMessage
func onMessage(client proto.BackupClient, msg RawMessage) {
	log.Println("Getting onMessage for point", msg.Data)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	log.Println(ctx)
	defer cancel()
	feature, err := client.OnMessageIn(ctx, &proto.Chat{
		MessageId: msg.MessageID,
		RoomId:    msg.RoomID,
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

func isReady(client proto.BackupClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ok, err := client.IsReady(ctx, &proto.Empty{})
	if err != nil {
		log.Fatalf("%v.isReady(_) = _, %v: ", client, err)
	}
	log.Println(ok)
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	client := proto.NewBackupClient(conn)
	log.Println(client)

	msg := RawMessage{}
	defer conn.Close()
	onMessage(client, msg)
	isReady(client)
}
