package main

import (
	"backendSenior/data/repository/chatsocket"
	be_mongo_repository "backendSenior/data/repository/mongo_repository"
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
	conn, err := mgo.Dial(utils.MONGO_CONN_STRING)
	if err != nil {
		log.Fatalln("Error connecting to mongo:", err)
	}

	// Repo
	roomUserRepo := delegate.NewDelegateRoomUserRepository(utils.CONTROLLER_ORIGIN)
	pool := chatsocket.NewConnectionPool()
	msgRepo := &be_mongo_repository.MessageRepositoryMongo{ConnectionDB: conn}

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
	downstreamService := service.NewChatDownstreamService(roomUserRepo, pool, pool, msgRepo, enc)
	upstreamService := service.NewChatUpstreamService(upstream, enc)
	delegateAuth := service.NewDelegateAuthService(utils.CONTROLLER_ORIGIN)

	router := (&route.RouterDeps{
		UpstreamService:   upstreamService,
		DownstreamService: downstreamService,
		AuthService:       delegateAuth,
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
