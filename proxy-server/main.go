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
	// Refactor :
	// Task: Plugin-Encryption : Check Flag to Forward
	// Add port to communicate with RPC
	enablePlugin := false
	if os.Getenv("PLUGIN_ACTIVE") == "True" {
		enablePlugin = true
	}

	enablePluginEnc := false
	if os.Getenv("PLUGIN_Encryption") == "True" {
		enablePluginEnc = true
	}

	pluginPort := os.Getenv("PLUGIN_PORT")
	if pluginPort == "" {
		enablePlugin = false
		fmt.Println("[NOTICE] Plugin is not enabled since PLUGIN_PATH is not set")
	}

	log.Println("Plugin Config >>>", "enablePlugin", enablePlugin, "enablePluginEnc", enablePluginEnc)
	log.Println("pluginPort", pluginPort)

	onMessagePlugin := plugin.NewOnMessagePortPlugin(enablePlugin, enablePluginEnc, pluginPort)
	upstream := upstream.NewUpStreamController(utils.CONTROLLER_ORIGIN, clientID, clientSecret)
	keystore := &mongo_repository.KeyRepository{}

	// err = onMessagePlugin.Wait()
	// if err != nil {
	// 	log.Fatalln("Wait for onMessagePlugin Error")
	// }
	// Task: Plugin-Encryption : Check Flag to Forward

	enc := service.NewEncryptionService(keystore, onMessagePlugin)
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
	defer onMessagePlugin.CloseConnection()
}
