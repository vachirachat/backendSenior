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
	"proxySenior/config"
	"proxySenior/controller/chat"
	"proxySenior/controller/route"
	"proxySenior/data/repository/delegate"
	"proxySenior/data/repository/mongo_repository"
	"proxySenior/data/repository/upstream"
	model_proxy "proxySenior/domain/model"
	"proxySenior/domain/plugin"
	"proxySenior/domain/service"
	"proxySenior/domain/service/key_service"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/joho/godotenv"

	_ "net/http/pprof"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Can't load .env file, does it exist ?")
	}

	// Still Hardcode
	connectionDB, err := mgo.Dial(config.MongoConnString)
	if err != nil {
		log.Panic("Can no connect Database", err.Error())
	}

	// Init Repository
	keystore := mongo_repository.NewKeyRepositoryMongo(connectionDB)

	// Repo
	roomUserRepo := delegate.NewDelegateRoomUserRepository(config.ControllerOrigin)
	pool := chatsocket.NewConnectionPool()
	msgRepo := delegate.NewDelegateMessageRepository(config.ControllerOrigin)
	proxyMasterAPI := delegate.NewRoomProxyAPI(config.ControllerOrigin)
	keyAPI := delegate.NewKeyAPI(config.ControllerOrigin)

	// Service
	clientID := config.ClientID
	if !bson.IsObjectIdHex(clientID) {
		log.Fatalln("error: please set valid CLIENT_ID")
	}
	clientSecret := config.ClientSecret
	if clientSecret == "" {
		log.Fatalln("error: please set client secret")
	}
	// Refactor :
	// Task: Plugin-Encryption : Check Flag to Forward
	// Add port to communicate with RPC
	var proxyConfig = &model_proxy.ProxyConfig{
		EnablePlugin:    false,
		EnablePluginEnc: false,
		PluginPort:      os.Getenv("PLUGIN_PORT"),
		DockerID:        "",
	}

	if os.Getenv("PLUGIN_ACTIVE") == "True" {
		proxyConfig.EnablePlugin = true
	}
	if os.Getenv("PLUGIN_Encryption") == "True" {
		proxyConfig.EnablePluginEnc = true
	}

	pluginPort := os.Getenv("PLUGIN_PORT")

	if pluginPort == "" {
		proxyConfig.EnablePlugin = false
		fmt.Println("[NOTICE] Plugin is not enabled since PLUGIN_PORT is not set")
	}

	onMessagePlugin := plugin.NewOnMessagePortPlugin(proxyConfig)

	upStreamController := upstream.NewUpStreamController(config.ControllerOrigin, clientID, clientSecret)
	defer upStreamController.Stop()
	upstreamService := service.NewChatUpstreamService(upStreamController)

	conn := make(chan struct{}, 10)
	upstreamService.OnConnect(conn)
	defer upstreamService.OffConnect(conn)

	rabbit := rmq.New(config.RabbitMQConnString)
	if err := rabbit.Connect(); err != nil {
		log.Fatalf("can't connect to rabbitmq %s\n", err)
	}
	for _, q := range []string{"upload_task", "upload_result"} {
		if err := rabbit.EnsureQueue(q); err != nil {
			log.Fatalf("error ensuring queue %s: %s\n", q, err)
		}
	}
	keyService := key_service.New(keystore, keyAPI, proxyMasterAPI, clientID)
	keyService.InitKeyPair()

	enc := service.NewEncryptionService(keyService, onMessagePlugin)

	downstreamService := service.NewChatDownstreamService(roomUserRepo, pool, pool, nil) // no message repo needed
	delegateAuth := service.NewDelegateAuthService(config.ControllerOrigin)

	messageService := service.NewMessageService(msgRepo)
	fileService := service.NewFileService(config.ControllerOrigin, rabbit)

	go func() {
		err := fileService.Run() // go routing waiting for message
		if err != nil {
			log.Println("file service: ", err)
		}
	}()

	configService := service.NewConfigService(enc, proxyConfig, onMessagePlugin)
	stickerService := service.NewStickerService(config.ControllerBasePath)
	// Fix Real Use
	// configService := service.NewConfigService(enc, proxyConfig)
	// create router from service
	router := (&route.RouterDeps{
		UpstreamService:   upstreamService,
		DownstreamService: downstreamService,
		AuthService:       delegateAuth,
		MessageService:    messageService,
		ConfigService:     configService,
		KeyService:        keyService,
		FileService:       fileService,
		Encrpytion:        enc,
		StickerService:    stickerService,
	}).NewRouter()

	router.GET("/debug/conns", func(c *gin.Context) {

		c.JSON(200, pool.DebugNumOfConns())
	})

	// websocket messasge handler
	messageHandler := chat.NewMessageHandler(upstreamService, downstreamService, roomUserRepo, keyService, onMessagePlugin, enc)
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
