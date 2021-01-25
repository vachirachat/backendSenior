package main

import (
	"backendSenior/data/repository/chatsocket"
	"log"
	"proxySenior/controller/chat"
	"proxySenior/controller/route"
	"proxySenior/data/repository/delegate"
	"proxySenior/data/repository/mongo_repository"
	"proxySenior/data/repository/upstream"
	"proxySenior/domain/plugin"
	"proxySenior/domain/service"
	"proxySenior/domain/service/key_service"
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
	roomUserRepo := delegate.NewDelegateRoomUserRepository(utils.ControllerOrigin)
	pool := chatsocket.NewConnectionPool()
	msgRepo := delegate.NewDelegateMessageRepository(utils.ControllerOrigin)
	proxyMasterAPI := delegate.NewRoomProxyAPI(utils.ControllerOrigin)
	keyAPI := delegate.NewKeyAPI(utils.ControllerOrigin)

	// Service
	clientID := utils.ClientID
	if !bson.IsObjectIdHex(clientID) {
		log.Fatalln("error: please set valid CLIENT_ID")
	}
	clientSecret := utils.ClientSecret
	if clientSecret == "" {
		log.Fatalln("error: please set client secret")
	}

	enablePlugin := true
	pluginPath := utils.PluginPath

	upstream := upstream.NewUpStreamController(utils.ControllerOrigin, clientID, clientSecret)
	defer upstream.Stop()
	upstreamService := service.NewChatUpstreamService(upstream)

	conn := make(chan struct{}, 10)
	upstreamService.OnConnect(conn)
	defer upstreamService.OffConnect(conn)

	onMessagePlugin := plugin.NewOnMessagePlugin(enablePlugin, pluginPath)

	err = onMessagePlugin.Wait()
	if err != nil {
		log.Fatalln("Wait for onMessagePlugin Error")
	}

	downstreamService := service.NewChatDownstreamService(roomUserRepo, pool, pool, nil) // no message repo needed
	delegateAuth := service.NewDelegateAuthService(utils.ControllerOrigin)
	keyService := key_service.New(keystore, keyAPI, proxyMasterAPI, clientID)
	keyService.InitKeyPair()

	messageService := service.NewMessageService(msgRepo)

	// create router from service
	router := (&route.RouterDeps{
		UpstreamService:   upstreamService,
		DownstreamService: downstreamService,
		AuthService:       delegateAuth,
		MessageService:    messageService,
		KeyService:        keyService,
	}).NewRouter()

	// websocket messasge handler
	messageHandler := chat.NewMessageHandler(upstreamService, downstreamService, roomUserRepo, keyService, onMessagePlugin)
	go messageHandler.Start()

	// TODO; refactor
	// auto reset key
	go func() {
		for {
			<-conn
			keyService.RevalidateAll()
		}
	}()

	router.Run(utils.ListenAddress)

}
