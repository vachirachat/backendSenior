package main

import (
	route "backendSenior/controller/handler"
	"backendSenior/data/repository/mongo_repository"
	service "backendSenior/domain/service"
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

	// Init service
	authSvc := &auth.AuthService{
		UserRepository: userRepo,
	}
	msgSvc := service.NewMessageService(messageRepo)
	userSvc := service.NewUserService(userRepo)
	roomSvc := service.NewRoomService(roomRepo)

	routerDeps := route.RouterDeps{
		RoomService:    roomSvc,
		MessageService: msgSvc,
		UserService:    userSvc,
		AuthService:    authSvc,
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
