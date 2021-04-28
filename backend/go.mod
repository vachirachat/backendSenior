module backendSenior

go 1.15

//replace proxySenior => ../proxy-server

replace common => ../common

replace backendSenior => ../backend

require (
	common v0.0.0-00010101000000-000000000000
	firebase.google.com/go/v4 v4.1.0
	github.com/ahmetb/go-linq/v3 v3.2.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.6.2
	github.com/gin-gonic/gin v1.6.3
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-ini/ini v1.62.0 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/minio/minio-go v6.0.14+incompatible
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/reactivex/rxgo/v2 v2.4.0
	github.com/satori/go.uuid v1.2.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	google.golang.org/api v0.30.0
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637 // indirect
)
