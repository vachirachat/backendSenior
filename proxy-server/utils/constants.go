package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func defaultEnv(key string, defaultVal string) string {
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	return defaultVal
}

func requiredEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("ERROR: environment variable %s is not set", key))
	}
	return val
}

var (
	// ControllerOrigin is contrller's IP + port
	ControllerOrigin string
	// ControllerBasePath is schene://host:port of controller
	ControllerBasePath string
	// MongoConnString is connection string mongo in form mongodb://host:port
	MongoConnString string
	// ListenAddress usually ":PORT"
	ListenAddress string
	// PProfAddress is used for profiling
	// see here for more info https://golang.org/pkg/net/http/pprof/
	PProfAddress string
	// ClientID is proxyID, created proxy at controller
	ClientID string
	// ClientSecret is proxySecret, returned when created proxy at controller
	ClientSecret string
	// PluginPath path to plugin file
	PluginPath string
	// RabbitMQConnString used for connecting rabbit mq
	RabbitMQConnString string
)

// setup env
func init() {
	envFile := defaultEnv("ENV_FILE", ".env")

	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalln("error loading env", err)
	}

	ControllerOrigin = requiredEnv("CONTROLLER_ORIGIN")
	ControllerBasePath = fmt.Sprintf("http://%s", ControllerOrigin)
	MongoConnString = requiredEnv("MONGO_CONN_STRING")
	RabbitMQConnString = requiredEnv("RABBITMQ_CONN_STRING")
	ListenAddress = defaultEnv("PORT", ":8090")
	PProfAddress = defaultEnv("PPROF_ADDRESS", "localhost:6060") // don't allow remote pprof by default
	ClientID = requiredEnv("CLIENT_ID")
	ClientSecret = requiredEnv("CLIENT_SECRET")
	// PluginPath = requiredEnv("PLUGIN_PATH")
}

var (
	PATH_ORIGIN             = "./share/temp_file/"
	PATH_ORIGIN_ZIP         = "./share/temp_zip/"
	DOCKER_PATH_ORIGIN      = "/app/go_server"
	DOCKEREXEC_FILE_NAME    = "docker_exec"
	DOCKERIMAGE_NAME        = "docker_upload"
	DOCKERIMAGE_REMOTE_NAME = "docker_upload"
	PROXY_KEY               = "0123456789abcdef"
)
