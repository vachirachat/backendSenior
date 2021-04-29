package main

// ROAD: I think this isn't used anymore

//var isReady bool
//var conn *mgo.Session
//
//type BackupServer struct {
//	proto.UnimplementedBackupServer
//	saveChats []*proto.Chat // read-only Chat after initialized
//}
//
//type BackupMessage struct {
//	MessageID bson.ObjectId `bson:"_id"`
//	TimeStamp time.Time     `bson:"timestamp"`
//	RoomID    bson.ObjectId `bson:"roomId"`
//	UserID    bson.ObjectId `bson:"userId"`
//	ClientUID string        `bson:"clientUID"`
//	Data      string        `bson:"data"`
//	Type      string        `bson:"type"`
//}
//
//func (b *BackupServer) OnMessageIn(context context.Context, chat *proto.Chat) (*proto.Empty, error) {
//	log.Println("Access OnMessageIn")
//	bMsg := BackupMessage{
//		MessageID: bson.ObjectIdHex(chat.MessageId),
//		TimeStamp: time.Unix(chat.Timestamp, 0),
//		RoomID:    bson.ObjectIdHex(chat.RoomId),
//		UserID:    bson.ObjectIdHex(chat.UserId),
//		ClientUID: chat.ClientUid,
//		Data:      chat.Data,
//		Type:      chat.Type,
//	}
//	log.Println("Incoming Message  >>>>>>", bMsg, "\n")
//	var message []BackupMessage
//	conn.DB("backup").C("message").Find(nil).All(&message)
//	for i, v := range message {
//		if i > len(message)-5 {
//			log.Println(v, "\n")
//		}
//
//	}
//	// return &proto.Empty{}, nil
//	return &proto.Empty{}, conn.DB("backup").C("message").Insert(bMsg)
//
//}
//func (b *BackupServer) IsReady(context context.Context, empty *proto.Empty) (*proto.Status, error) {
//	log.Println("Access IsReady")
//	log.Println("Access ", empty)
//	return &proto.Status{Ok: true}, nil
//}
//
//func NewBackupServer() proto.BackupServer {
//	return &BackupServer{saveChats: make([]*proto.Chat, 0)}
//}
//
//func main() {
//	//connect mongo Server
//
//	go func() {
//		var err error
//		conn, err = mgo.Dial("172.17.0.2:27017")
//		if err != nil {
//			log.Fatal("error running", err)
//		}
//		isReady = true
//	}()
//
//	lis, err := net.Listen("tcp", fmt.Sprintf("172.17.0.3:%d", 5005))
//	if err != nil {
//		log.Fatalf("failed to listen: %v", err)
//	}
//	var opts []grpc.ServerOption
//	opts = []grpc.ServerOption{}
//	log.Println("Start-go Server with PORT", "5005")
//	grpcServer := grpc.NewServer(opts...)
//	proto.RegisterBackupServer(grpcServer, NewBackupServer())
//	grpcServer.Serve(lis)
//}
//
//var Keymap = = "abcdefghijklmnopqrstuvwxyz012345"
//// Decrypt takes a message, then return message with data decrypted with appropiate key
//func encryptedMessage(ctx context.Context, in *proto.Chat, opts ...grpc.CallOption) (*proto.Chat, error) {
//	message := BackupMessage{
//		MessageID: bson.ObjectIdHex(chat.MessageId),
//		TimeStamp: time.Unix(chat.Timestamp, 0),
//		RoomID:    bson.ObjectIdHex(chat.RoomId),
//		UserID:    bson.ObjectIdHex(chat.UserId),
//		ClientUID: chat.ClientUid,
//		Data:      chat.Data,
//		Type:      chat.Type,
//	}
//
//	key := Keymap
//
//	// b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(message.Data)))
//	cipherText, err := base64.StdEncoding.DecodeString(message.Data)
//	if err != nil {
//		return message, fmt.Errorf("decode b64: %s", err.Error())
//	}
//
//	block, err := aes.NewCipher(key)
//	if err != nil {
//		return message, err
//	}
//
//	if len(cipherText) < aes.BlockSize {
//		return message, errors.New("cipher text too short")
//	}
//
//	iv := cipherText[:aes.BlockSize]
//	data := cipherText[aes.BlockSize:]
//	if len(data)%aes.BlockSize != 0 {
//		return message, errors.New("wrong cipher text size")
//	}
//
//	mode := cipher.NewCBCDecrypter(block, iv)
//	if err != nil {
//		return message, err
//	}
//
//	decrypted := make([]byte, len(data))
//
//	mode.CryptBlocks(decrypted, data)
//
//	decrypted, _ = pkcs7.Unpad(decrypted, aes.BlockSize)
//	message.Data = string(decrypted)
//
//	return message, nil
//}
