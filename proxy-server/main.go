package main

import (
	"backendSenior/data/repository/chatsocket"
	"fmt"
	"log"
	"os"
	"proxySenior/controller/chat"
	"proxySenior/controller/route"
	"proxySenior/data/repository/delegate"
	"proxySenior/data/repository/mongo_repository"
	"proxySenior/data/repository/upstream"
	"proxySenior/domain/plugin"
	"proxySenior/domain/service"
	"proxySenior/utils"

	"github.com/joho/godotenv"

	"github.com/globalsign/mgo/bson"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Can't load .env file, does it exist ?")
	}

	// Repo
	roomUserRepo := delegate.NewDelegateRoomUserRepository(utils.CONTROLLER_ORIGIN)
	pool := chatsocket.NewConnectionPool()
	msgRepo := delegate.NewDelegateMessageRepository(utils.CONTROLLER_ORIGIN)

	// Service
	clientID := os.Getenv("CLIENT_ID")
	if !bson.IsObjectIdHex(clientID) {
		log.Fatalln("error: please set valid CLIENT_ID")
	}
	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatalln("error: please set client secret")
	}

	// Add port to communicate with RPC
	enablePlugin := true
	pluginPort := os.Getenv("PLUGIN_PORT")
	if pluginPort == "" {
		enablePlugin = false
		fmt.Println("[NOTICE] Plugin is not enabled since PLUGIN_PATH is not set")
	}

	onMessagePlugin := plugin.NewOnMessagePortPlugin(enablePlugin, pluginPort)
	defer onMessagePlugin.CloseConnection()

	upstream := upstream.NewUpStreamController(utils.CONTROLLER_ORIGIN, clientID, clientSecret)
	keystore := &mongo_repository.KeyRepository{}

	err = onMessagePlugin.Wait()
	if err != nil {
		log.Fatalln("Wait for onMessagePlugin Error")
	}

	// var message model.Message
	// message = model.Message{
	// 	MessageID: bson.NewObjectId(),
	// 	TimeStamp: time.Now(),
	// 	RoomID:    bson.NewObjectId(),
	// 	UserID:    bson.NewObjectId(),
	// 	ClientUID: "waritphon",
	// 	Data:      "Test - data",
	// 	Type:      "TEST",
	// }
	// err = external.OnMessageIn(message)

	enc := service.NewEncryptionService(keystore)
	downstreamService := service.NewChatDownstreamService(roomUserRepo, pool, pool, nil) // no message repo needed
	upstreamService := service.NewChatUpstreamService(upstream, enc)
	delegateAuth := service.NewDelegateAuthService(utils.CONTROLLER_ORIGIN)
	messageService := service.NewMessageService(msgRepo, enc)

	// create router from service
	router := (&route.RouterDeps{
		UpstreamService:   upstreamService,
		DownstreamService: downstreamService,
		AuthService:       delegateAuth,
		MessageService:    messageService,
	}).NewRouter()

	// websocket messasge handler
	messageHandler := chat.NewMessageHandler(upstreamService, downstreamService, roomUserRepo, enc, onMessagePlugin)
	go messageHandler.Start()

	router.Run(utils.LISTEN_ADDRESS)

}
