package main

import (
	"backendSenior/data/repository/chatsocket"
	"backendSenior/domain/model"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"proxySenior/controller/route"
	"proxySenior/data/repository/delegate"
	"proxySenior/data/repository/mongo_repository"
	"proxySenior/data/repository/upstream"
	"proxySenior/domain/service"
	"proxySenior/utils"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func main() {
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

	enc := service.NewEncryptionService(keystore)
	downstreamService := service.NewChatDownstreamService(roomUserRepo, pool, pool, nil, enc)
	upstreamService := service.NewChatUpstreamService(upstream, enc)
	delegateAuth := service.NewDelegateAuthService(utils.CONTROLLER_ORIGIN)
	messageService := service.NewMessageService(msgRepo, enc)

	router := (&route.RouterDeps{
		UpstreamService:   upstreamService,
		DownstreamService: downstreamService,
		AuthService:       delegateAuth,
		MessageService:    messageService,
	}).NewRouter()

	pipe := make(chan []byte, 100)
	upstreamService.RegsiterHandler(pipe)

	go func() {
		for {
			data := <-pipe
			fmt.Printf("[upstream] <-- %s\n", data)
			var msg model.Message
			err := json.Unmarshal(data, &msg)
			if err != nil {
				fmt.Println("Error unmarshal:", err)
				continue
			}
			fmt.Println("The message is", msg)

			err = downstreamService.BroadcastMessageToRoom(msg.RoomID.Hex(), msg)
			if err != nil {
				fmt.Println("Error BCasting", err)
			}
		}
	}()

	router.Run(utils.LISTEN_ADDRESS)

}
