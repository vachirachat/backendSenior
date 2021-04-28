package utills

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
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

// Server const
var (
	MongoEndpoint string
	ListenAddress string

	MinioEndpoint  string
	MinioAccessID  string
	MinioSecretKey string

	JWTSecret        string
	JWTRefreshSecret string

	PProfAddress string
)

func init() {
	envFile := defaultEnv("ENV_FILE", ".env")

	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalln("error loading env", err)
	}

	MongoEndpoint = requiredEnv("MONGO_ENDPOINT")
	ListenAddress = defaultEnv("LISTEN_ADDR", ":8080")

	MinioEndpoint = requiredEnv("MINIO_ENDPOINT")
	MinioAccessID = requiredEnv("MINIO_ACCESS_ID")
	MinioSecretKey = requiredEnv("MINIO_SECRET_KEY")

	JWTSecret = requiredEnv("JWT_SECRET")
	JWTRefreshSecret = requiredEnv("JWT_REFRESH_SECRET")

	PProfAddress = defaultEnv("PPROF_ADDR", "localhost:6061")

}

// Role Const
const (
	ROLEADMIN = "admin"
	ROLEUSER  = "user"
)

const ADMIN_MODE = true
