module proxySenior

go 1.15

replace backendSenior => ../backend

replace common => ../common

replace proxySenior => ../proxy-server

require (
	backendSenior v0.0.0-00010101000000-000000000000
	common v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.6.3
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-resty/resty/v2 v2.5.0
	github.com/golang/protobuf v1.4.2
	github.com/gorilla/websocket v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/mergermarket/go-pkcs7 v0.0.0-20170926155232-153b18ea13c9
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.25.0
)
