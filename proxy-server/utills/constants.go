package utills

var (
	CONTROLLER_ORIGIN       = "localhost:8080"
	PATH_ORIGIN_PROXY       = "./share/"
	PATH_ORIGIN             = "./share/temp_file/"
	PATH_ORIGIN_ZIP         = "./share/temp_zip/"
	DOCKER_PATH_ORIGIN      = "/app/go_server"
	MONGO_CONN_STRING       = "mongodb://localhost:27017"
	DOCKEREXEC_FILE_NAME    = "docker_exec"
	DOCKERIMAGE_NAME        = "docker_upload"
	DOCKERIMAGE_REMOTE_NAME = "docker_upload"
	LISTEN_ADDRESS          = defaultEnv("PORT", ":8090")
	PROXY_KEY               = "0123456789abcdef"
)
