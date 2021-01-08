module proxySenior

go 1.15

replace backendSenior => ../backend

require (
	backendSenior v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.3.0
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/golang/protobuf v1.4.2
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-plugin v1.4.0
	github.com/mergermarket/go-pkcs7 v0.0.0-20170926155232-153b18ea13c9
	google.golang.org/grpc v1.31.0
)
