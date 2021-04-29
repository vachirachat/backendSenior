module proxySenior

go 1.15

replace backendSenior => ../backend

replace common => ../common

replace proxySenior => ../proxy-server

replace go-module => ../plugin/go-module

require (
	backendSenior v0.0.0-00010101000000-000000000000
	common v0.0.0-00010101000000-000000000000
	github.com/cornelk/hashmap v1.0.1
	github.com/fsouza/go-dockerclient v1.7.2
	github.com/gin-gonic/gin v1.6.3
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-resty/resty/v2 v2.5.0
	github.com/golang/protobuf v1.4.2
	github.com/gorilla/websocket v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/mergermarket/go-pkcs7 v0.0.0-20170926155232-153b18ea13c9
	go-module v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0

)
