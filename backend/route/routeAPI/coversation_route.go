package routeAPI

import (
	repository "backendSenior/repository/pubsub-chat"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/websocket"
)

func AddCoversationRoute(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	//Init new hub server
	hub := repository.NewHub()
	go hub.Run()

	routerGroup.GET("/ws", func(context *gin.Context) {
		fmt.Println("new connection!")
		w := context.Writer
		r := context.Request

		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(req *http.Request) bool {
				return true
			},
		}

		// Reading username from request parameter
		username := r.URL.Query()
		userID := username.Get("userID")
		log.Println(username)
		// Upgrading the HTTP connection socket connection
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		repository.CreateNewSocketUser(hub, connection, bson.ObjectIdHex(userID))

	})

}
