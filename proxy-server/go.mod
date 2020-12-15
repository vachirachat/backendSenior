module proxySenior

go 1.15

replace backendSenior => ../backend

require (
	backendSenior v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.3.0
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/gorilla/websocket v1.4.2
)
