package main

import (
	"backendSenior/data/repository/chatsocket"
	"log"
	"os"
	"proxySenior/controller/chat"
	"proxySenior/controller/route"
	"proxySenior/data/repository/delegate"
	"proxySenior/data/repository/mongo_repository"
	"proxySenior/data/repository/upstream"
	"proxySenior/domain/service"
	"proxySenior/utils"

	"github.com/globalsign/mgo/bson"
)

func main() {
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

	upstream := upstream.NewUpStreamController(utils.CONTROLLER_ORIGIN, clientID, clientSecret)
	keystore := &mongo_repository.KeyRepository{}

	enc := service.NewEncryptionService(keystore)
	downstreamService := service.NewChatDownstreamService(roomUserRepo, pool, pool, nil, enc)
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
	messageHandler := chat.NewMessageHandler(upstreamService, downstreamService, roomUserRepo)
	go messageHandler.Start()

	router.Run(utils.LISTEN_ADDRESS)

}
