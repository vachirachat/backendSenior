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

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Can't load .env file, does it exist ?")
	}

	// Still Hardcode
	connectionDB, err := mgo.Dial("mongodb://localhost:27017")
	if err != nil {
		log.Panic("Can no connect Database", err.Error())
	}

	// Init Repository
	keystore := mongo_repository.NewKeyRepositoryMongo(connectionDB)

	// Repo
	roomUserRepo := delegate.NewDelegateRoomUserRepository(utils.CONTROLLER_ORIGIN)
	pool := chatsocket.NewConnectionPool()
	msgRepo := delegate.NewDelegateMessageRepository(utils.CONTROLLER_ORIGIN)
	proxyMasterAPI := delegate.NewRoomProxyAPI(utils.CONTROLLER_ORIGIN)

	// Service
	clientID := os.Getenv("CLIENT_ID")
	if !bson.IsObjectIdHex(clientID) {
		log.Fatalln("error: please set valid CLIENT_ID")
	}
	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatalln("error: please set client secret")
	}

	enablePlugin := true
	pluginPath := os.Getenv("PLUGIN_PATH")
	if pluginPath == "" {
		enablePlugin = false
		fmt.Println("[NOTICE] Plugin is not enabled since PLUGIN_PATH is not set")
	}

	upstream := upstream.NewUpStreamController(utils.CONTROLLER_ORIGIN, clientID, clientSecret)
	onMessagePlugin := plugin.NewOnMessagePlugin(enablePlugin, pluginPath)

	err = onMessagePlugin.Wait()
	if err != nil {
		log.Fatalln("Wait for onMessagePlugin Error")
	}

	enc := service.NewEncryptionService(keystore)
	downstreamService := service.NewChatDownstreamService(roomUserRepo, pool, pool, nil) // no message repo needed
	upstreamService := service.NewChatUpstreamService(upstream, enc)
	delegateAuth := service.NewDelegateAuthService(utils.CONTROLLER_ORIGIN)
	messageService := service.NewMessageService(msgRepo, enc)
	keyService := service.NewKeyService(keystore, proxyMasterAPI, clientID)

	// create router from service
	router := (&route.RouterDeps{
		UpstreamService:   upstreamService,
		DownstreamService: downstreamService,
		AuthService:       delegateAuth,
		MessageService:    messageService,
		KeyService:        keyService,
	}).NewRouter()

	// websocket messasge handler
	messageHandler := chat.NewMessageHandler(upstreamService, downstreamService, roomUserRepo, enc, onMessagePlugin)
	go messageHandler.Start()

	router.Run(utils.LISTEN_ADDRESS)

}
