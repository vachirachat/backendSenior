package main

import (
	"backendSenior/data/repository/chatsocket"
	"common/rmq"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"proxySenior/controller/chat"
	"proxySenior/controller/route"
	"proxySenior/data/repository/delegate"
	"proxySenior/data/repository/mongo_repository"
	"proxySenior/data/repository/upstream"
	"proxySenior/domain/plugin"
	"proxySenior/domain/service"
	"proxySenior/domain/service/key_service"
	"proxySenior/utils"
	"syscall"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/joho/godotenv"
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
	// Refactor :
	// Task: Plugin-Encryption : Check Flag to Forward
	// Add port to communicate with RPC
	enablePlugin := false
	if os.Getenv("PLUGIN_ACTIVE") == "True" {
		enablePlugin = true
	}

	enablePluginEnc := false
	if os.Getenv("PLUGIN_Encryption") == "True" {
		enablePluginEnc = true
	}

	pluginPort := os.Getenv("PLUGIN_PORT")
	log.Println("pluginPort =", pluginPort)

	if pluginPort == "" {
		enablePlugin = false
		fmt.Println("[NOTICE] Plugin is not enabled since PLUGIN_PORT is not set")
	}

	log.Println("Plugin Config >>>", "enablePlugin", enablePlugin, "enablePluginEnc", enablePluginEnc)
	log.Println("pluginPort", pluginPort)

	onMessagePlugin := plugin.NewOnMessagePortPlugin(enablePlugin, enablePluginEnc, pluginPort)

	upstream := upstream.NewUpStreamController(utils.ControllerOrigin, clientID, clientSecret)
	defer upstream.Stop()
	upstreamService := service.NewChatUpstreamService(upstream)

	conn := make(chan struct{}, 10)
	upstreamService.OnConnect(conn)
	defer upstreamService.OffConnect(conn)

	if enablePlugin {
		log.Println("waiting plugin")
		err = onMessagePlugin.Wait()
		if err != nil {
			log.Fatalln("Wait for onMessagePlugin Error")
		}

	}

	rabbit := rmq.New("amqp://guest:guest@localhost:5672/")
	if err := rabbit.Connect(); err != nil {
		log.Fatalf("can't connect to rabbitmq %s\n", err)
	}
	for _, q := range []string{"upload_task", "upload_result"} {
		if err := rabbit.EnsureQueue(q); err != nil {
			log.Fatalf("error ensuring queue %s: %s\n", q, err)
		}
	}
	enc := service.NewEncryptionService(keystore, onMessagePlugin)

	downstreamService := service.NewChatDownstreamService(roomUserRepo, pool, pool, nil) // no message repo needed
	delegateAuth := service.NewDelegateAuthService(utils.ControllerOrigin)
	keyService := key_service.New(keystore, keyAPI, proxyMasterAPI, clientID)
	keyService.InitKeyPair()

	messageService := service.NewMessageService(msgRepo)
	fileService := service.NewFileService("localhost:8080", rabbit)

	go fileService.Run() // go routing waiting for message

	// create router from service
	router := (&route.RouterDeps{
		UpstreamService:   upstreamService,
		DownstreamService: downstreamService,
		AuthService:       delegateAuth,
		MessageService:    messageService,
		KeyService:        keyService,
		FileService:       fileService,
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

	httpSrv := &http.Server{
		Addr:    utils.ListenAddress,
		Handler: router,
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("Listen: %s\n", err)
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

	log.Println("Server exiting")

}
