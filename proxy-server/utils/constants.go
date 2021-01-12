package utils

import (
	"os"

	"github.com/joho/godotenv"
)

func defaultEnv(key string, defaultVal string) string {
	// TODO: handle error
	_ = godotenv.Load()

	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	return defaultVal
}

var (
	CONTROLLER_ORIGIN = "localhost:8080"
	MONGO_CONN_STRING = "mongodb://localhost:27017"
	LISTEN_ADDRESS    = defaultEnv("PORT", ":8090")
)

const BACKKOKTIMEZONE = "Asia/Bangkok"
