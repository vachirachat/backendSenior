package main

import (
	route "backendSenior/controller/handler"
	"backendSenior/data/repository/chatsocket"
	"backendSenior/data/repository/mongo_repository"
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"
	"backendSenior/utills"
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"github.com/globalsign/mgo"
	"google.golang.org/api/option"
)

var app *firebase.App

func initFirebase() {
	opt := option.WithCredentialsFile("../../account-secret-key.json")
	config := &firebase.Config{ProjectID: "senior-project-mychat"}
	var err error
	app, err = firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Fatalf("[Firebase] error init firebase app %s\n", err)
	} else {
		log.Println("[Firebase] successfully connected to firebase app")
	}
}

func main() {
	initFirebase()

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

	organizeRepo := mongo_repository.NewOrganizeRepositoryMongo(connectionDB)
	organizeUserRepo := mongo_repository.NewOrganizeUserRepositoryMongo(connectionDB)

	proxyRepo := mongo_repository.NewProxyRepositoryMongo(connectionDB)

	chatPool := chatsocket.NewConnectionPool()

	roomUserRepo := mongo_repository.NewCachedRoomUserRepository(connectionDB)
	roomProxyRepo := mongo_repository.NewCachedRoomProxyRepository(connectionDB)

	// Init service

	// TODO: implement token repo, no hardcode secret
	jwtSvc := auth.NewJWTService((repository.TokenRepository)(nil), []byte("secret_access"), []byte("secret_refresh"))

	msgSvc := service.NewMessageService(messageRepo)
	userSvc := service.NewUserService(userRepo, jwtSvc)
	roomSvc := service.NewRoomService(roomRepo, roomUserRepo, roomProxyRepo)
	organizeSvc := service.NewOrganizeService(organizeRepo, organizeUserRepo)
	// we use room proxy repo to map!
	chatSvc := service.NewChatService(roomProxyRepo, chatPool, chatPool, messageRepo)
	proxySvc := service.NewProxyService(proxyRepo)
	proxyAuthSvc := auth.NewProxyAuth(proxyRepo)

	routerDeps := route.RouterDeps{
		RoomService:      roomSvc,
		MessageService:   msgSvc,
		UserService:      userSvc,
		JWTService:       jwtSvc,
		ChatService:      chatSvc,
		ProxyService:     proxySvc,
		ProxyAuth:        proxyAuthSvc,
		OraganizeService: organizeSvc,
	}

	router := routerDeps.NewRouter()

	router.Run(utills.PORTWEBSERVER)
}
