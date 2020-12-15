package main

import (
	"backendSenior/data/repository/chatsocket"
	be_mongo_repository "backendSenior/data/repository/mongo_repository"
	"backendSenior/domain/model"
	"encoding/json"
	"fmt"
	"log"
	"proxySenior/controller/route"
	"proxySenior/data/repository/mongo_repository"
	"proxySenior/data/repository/upstream"
	"proxySenior/domain/service"
	"proxySenior/utils"

	"github.com/globalsign/mgo"
)

func main() {
	conn, err := mgo.Dial(utils.MONGO_CONN_STRING)
	if err != nil {
		log.Fatalln("Error connecting to mongo:", err)
	}

	roomUserRepo := mongo_repository.NewCachedRoomUserRepository(conn)
	pool := chatsocket.NewConnectionPool()
	msgRepo := &be_mongo_repository.MessageRepositoryMongo{ConnectionDB: conn}
	upstream := upstream.NewUpStreamController(utils.CONTROLLER_ORIGIN)

	enc := &service.EncryptionService{}
	downstreamService := service.NewChatDownstreamService(roomUserRepo, pool, pool, msgRepo, enc)
	upstreamService := service.NewChatUpstreamService(upstream, enc)

	router := (&route.RouterDeps{
		UpstreamService:   upstreamService,
		DownstreamService: downstreamService,
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
