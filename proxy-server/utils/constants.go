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
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("error loading env", err)
	}

	ControllerOrigin = "localhost:8080"
	MongoConnString = "mongodb://localhost:27017"
	ListenAddress = defaultEnv("PORT", ":8090")
	ClientID = requiredEnv("CLIENT_ID")
	ClientSecret = requiredEnv("CLIENT_SECRET")
	PluginPath = requiredEnv("PLUGIN_PATH")
}
