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
	// ClientID is proxyID, created proxy at controller
	ClientID string
	// ClientSecret is proxySecret, returned when created proxy at controller
	ClientSecret string
	// PluginPath path to plugin file
	PluginPath string
)

// setup env
func init() {
	envFile := defaultEnv("ENV_FILE", ".env")

	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalln("error loading env", err)
	}

	ControllerOrigin = "localhost:8080"
	ControllerBasePath = "http://localhost:8080"
	MongoConnString = "mongodb://localhost:27017"
	ListenAddress = defaultEnv("PORT", ":8090")
	ClientID = requiredEnv("CLIENT_ID")
	ClientSecret = requiredEnv("CLIENT_SECRET")
	// PluginPath = requiredEnv("PLUGIN_PATH")
}

var (
	CONTROLLER_ORIGIN       = "localhost:8080"
	PATH_ORIGIN             = "./share/temp_file/"
	PATH_ORIGIN_ZIP         = "./share/temp_zip/"
	DOCKER_PATH_ORIGIN      = "/app/go_server"
	MONGO_CONN_STRING       = "mongodb://localhost:27017"
	DOCKEREXEC_FILE_NAME    = "docker_exec"
	DOCKERIMAGE_NAME        = "docker_upload"
	DOCKERIMAGE_REMOTE_NAME = "docker_upload"
	LISTEN_ADDRESS          = defaultEnv("PORT", ":8090")
	PROXY_KEY               = "0123456789abcdef"
	PATH_ORIGIN_PROXY       = "./share/"
)
