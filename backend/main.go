package main

import (
	"backendSenior/config"
	route "backendSenior/controller/handler"
	"backendSenior/data/repository/chatsocket"
	"backendSenior/data/repository/file"
	"backendSenior/data/repository/mongo_repository"
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/globalsign/mgo"
	"google.golang.org/api/option"
	_ "net/http/pprof"
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

	// clnt, err := app.Messaging(context.Background())
	// if err != nil {
	// 	log.Fatalln("error getting messaiging instance", err)
	// }

	// clnt.SendMulticast(context.Background(), &messaging.MulticastMessage{
	// 	Tokens: []string{"ckYOMA85QDC97cW4vqCVUn:APA91bFzE_i6_ZjVsMT78cLTeIPmWBiaiMuk8kOaVULuyKp_dJ1EhYk8_GJJEhBZnDUmvtU-DYXcEXLTnUwUj1uuR2yPZSSwb07AOeC3DtRnWkx5SDTNIVWTxNdX6xPpsQ1oqVUxieQZ"},
	// 	Data: map[string]string{
	// 		"foo": "bar",
	// 		"baz": "quax",
	// 	},
	// 	Notification: &messaging.Notification{
	// 		Title:    "This is test notification",
	// 		Body:     "Test from Golang",
	// 		ImageURL: "https://pkg.go.dev/static/img/go-logo-blue.svg",
	// 	},
	// })
}

func main() {
	initFirebase()

	connectionDB, err := mgo.Dial(config.MongoEndpoint)
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
	orgRoomRepo := mongo_repository.NewOrgRoomRepository(connectionDB)

	proxyRepo := mongo_repository.NewProxyRepositoryMongo(connectionDB)

	chatPool := chatsocket.NewConnectionPool()

	roomUserRepo := mongo_repository.NewCachedRoomUserRepository(connectionDB)
	roomProxyRepo := mongo_repository.NewCachedRoomProxyRepository(connectionDB)

	fcmTokenRepo := mongo_repository.NewFCMTokenRepository(connectionDB)
	fcmUserRepo := mongo_repository.NewFCMUserRepository(connectionDB)

	tokenRepo := mongo_repository.NewTokenRepository(connectionDB)
	fileStore, err := file.NewFileStore(&file.MinioConfig{
		Endpoint:  config.MinioEndpoint,
		AccessID:  config.MinioAccessID,
		SecretKey: config.MinioSecretKey,
		UseSSL:    false,
	})
	stickerRepo := mongo_repository.NewStickerRepository(connectionDB)

	if err != nil {
		log.Fatal("error creating fileStore:", err)
	}
	if err = fileStore.Init(); err != nil {
		log.Fatal("error init fileStore:", err)
	}

	fileMetaRepo := mongo_repository.NewFileMetaRepositoryMongo(connectionDB)

	clnt, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalln("Error getting messaging instance", err)
	}

	// Init service

	// TODO: implement token repo, no hardcode secret
	jwtSvc := auth.NewJWTService(tokenRepo, []byte(config.JWTSecret), []byte(config.JWTRefreshSecret))

	msgSvc := service.NewMessageService(messageRepo)
	userSvc := service.NewUserService(userRepo, jwtSvc)
	roomSvc := service.NewRoomService(roomRepo, roomUserRepo, roomProxyRepo)
	organizeSvc := service.NewOrganizeService(organizeRepo, organizeUserRepo, orgRoomRepo)
	notifSvc := service.NewNotificationService(fcmTokenRepo, fcmUserRepo, clnt)
	// we use room proxy repo to map!

	chatSvc := service.NewChatService(roomProxyRepo, roomUserRepo, chatPool, chatPool, messageRepo, notifSvc)

	proxySvc := service.NewProxyService(proxyRepo)
	proxyAuthSvc := auth.NewProxyAuth(proxyRepo)
	keyExSvc := service.NewKeyExchangeService(mongo_repository.KeyVersionCollection(connectionDB))

	fileSvc := service.NewFileService(fileStore, fileMetaRepo)
	stickerSvc := service.NewStickerService(stickerRepo, stickerRepo, fileStore)

	routerDeps := route.RouterDeps{
		RoomService:         roomSvc,
		MessageService:      msgSvc,
		UserService:         userSvc,
		JWTService:          jwtSvc,
		ChatService:         chatSvc,
		ProxyService:        proxySvc,
		ProxyAuth:           proxyAuthSvc,
		OraganizeService:    organizeSvc,
		NotificationService: notifSvc,
		KeyExchangeService:  keyExSvc,
		FileService:         fileSvc,
		StickerService:      stickerSvc,
	}

	router := routerDeps.NewRouter()

	httpSrv := &http.Server{
		Addr:    config.ListenAddress,
		Handler: router,
	}

	pprofServer := &http.Server{
		Addr:    config.PProfAddress,
		Handler: nil,
	}

	go func() {

		if err := pprofServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("pprof: Listen: %s\n", err)
		}
	}()

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("main: Listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	if err := pprofServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
