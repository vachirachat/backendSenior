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
	model_proxy "proxySenior/domain/model"
	"proxySenior/domain/plugin"
	"proxySenior/domain/service"
	"proxySenior/utills"

	"github.com/joho/godotenv"

	"github.com/globalsign/mgo/bson"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Can't load .env file, does it exist ?")
	}

	// Repo
	roomUserRepo := delegate.NewDelegateRoomUserRepository(utills.CONTROLLER_ORIGIN)
	pool := chatsocket.NewConnectionPool()
	msgRepo := delegate.NewDelegateMessageRepository(utills.CONTROLLER_ORIGIN)

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
	var proxyConfig = &model_proxy.ProxyConfig{
		EnablePlugin:    false,
		EnablePluginEnc: false,
		PluginPort:      os.Getenv("PLUGIN_PORT"),
		DockerID:        "08392baafeb2",
	}

	if os.Getenv("PLUGIN_ACTIVE") == "True" {
		proxyConfig.EnablePlugin = true
	}
	if os.Getenv("PLUGIN_Encryption") == "True" {
		proxyConfig.EnablePluginEnc = true
	}

	if proxyConfig.PluginPort == "" {
		proxyConfig.EnablePlugin = false
		fmt.Println("[NOTICE] Plugin is not enabled since PLUGIN_PATH is not set")
	}

	onMessagePlugin := plugin.NewOnMessagePortPlugin(proxyConfig)
	upstream := upstream.NewUpStreamController(utills.CONTROLLER_ORIGIN, clientID, clientSecret)
	keystore := &mongo_repository.KeyRepository{}

	enc := service.NewEncryptionService(keystore, onMessagePlugin)
	downstreamService := service.NewChatDownstreamService(roomUserRepo, pool, pool, nil) // no message repo needed
	upstreamService := service.NewChatUpstreamService(upstream, enc)
	delegateAuth := service.NewDelegateAuthService(utills.CONTROLLER_ORIGIN)
	messageService := service.NewMessageService(msgRepo, enc)

	configService := service.NewConfigService(enc, proxyConfig, onMessagePlugin)
	// Fix Real Use
	// configService := service.NewConfigService(enc, proxyConfig)
	// create router from service
	router := (&route.RouterDeps{
		UpstreamService:   upstreamService,
		DownstreamService: downstreamService,
		AuthService:       delegateAuth,
		MessageService:    messageService,
		ConfigService:     configService,
	}).NewRouter()

	// websocket messasge handler
	messageHandler := chat.NewMessageHandler(upstreamService, downstreamService, roomUserRepo, enc, onMessagePlugin)
	go messageHandler.Start()

	router.Run(utills.LISTEN_ADDRESS)
	defer onMessagePlugin.CloseConnection()
}
