package main

import (
	route "backendSenior/controller/handler"
	"backendSenior/data/repository/chatsocket"
	"backendSenior/data/repository/mongo_repository"
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"
	"backendSenior/utills"
	"log"

	"github.com/globalsign/mgo"
)

func main() {
	connectionDB, err := mgo.Dial(utills.MONGOENDPOINT)
	if err != nil {
		log.Panic("Can no connect Database", err.Error())
	}

	// Init Repository
	messageRepo := &mongo_repository.MessageRepositoryMongo{
		ConnectionDB: connectionDB,
	}

	userRepo := &mongo_repository.UserRepositoryMongo{
		ConnectionDB: connectionDB,
	}

	roomRepo := &mongo_repository.RoomRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	proxyRepo := mongo_repository.NewProxyRepositoryMongo(connectionDB)

	chatPool := chatsocket.NewConnectionPool()

	roomUserRepo := mongo_repository.NewCachedRoomUserRepository(connectionDB)
	roomProxyRepo := mongo_repository.NewCachedRoomProxyRepository(connectionDB)

	// Init service

	// TODO: implement token repo, no hardcode secret
	jwtSvc := auth.NewJWTService((repository.TokenRepository)(nil), []byte("secret_access"), []byte("secret_refresh"), make(map[string]bool))

	msgSvc := service.NewMessageService(messageRepo)
	userSvc := service.NewUserService(userRepo, jwtSvc)
	roomSvc := service.NewRoomService(roomRepo, roomUserRepo, roomProxyRepo)
	// we use room proxy repo to map!
	chatSvc := service.NewChatService(roomProxyRepo, chatPool, chatPool, messageRepo)
	proxySvc := service.NewProxyService(proxyRepo)
	proxyAuthSvc := auth.NewProxyAuth(proxyRepo)

	routerDeps := route.RouterDeps{
		RoomService:    roomSvc,
		MessageService: msgSvc,
		UserService:    userSvc,
		JWTService:     jwtSvc,
		ChatService:    chatSvc,
		ProxyService:   proxySvc,
		ProxyAuth:      proxyAuthSvc,
	}

	router := routerDeps.NewRouter()

	router.Run(utills.PORTWEBSERVER)

}

// func serveDefault(w http.ResponseWriter, r *http.Request) {
// 	log.Println(r.URL)
// 	if r.URL.Path != "/" {
// 		http.Error(w, "Not found", http.StatusNotFound)
// 		return
// 	}
// 	if r.Method != "GET" {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	http.ServeFile(w, r, "index.html")
// }

// func main() {
// 	hub := H
// 	go hub.Run()
// 	http.HandleFunc("/", serveDefault)
// 	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
// 		ServeWs(w, r)
// 	})
// 	//Listerning on port :8080...
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }
