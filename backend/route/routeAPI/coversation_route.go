package routeAPI

import (
	"backendSenior/api"
	socket "backendSenior/repository/pubsub-chat"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/gorilla/websocket"
)

func AddCoversationRoute(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	hub := socket.NewHub()
	go hub.Run()

	routerGroup.StaticFS("./public", http.Dir("./public/"))
	routerGroup.GET("/", api.RenderHome)

	routerGroup.GET("/ws", func(context *gin.Context) {
		w := context.Writer
		r := context.Request

		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		// Reading username from request parameter
		username := r.URL.Query()
		name := username.Get("nameid")
		log.Println(username)
		// Upgrading the HTTP connection socket connection
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		// FIX Room Query DATA
		// Query - Client map-to-Room
		room := []string


		// Fetch All Room Message - Client


		socket.CreateNewSocketUser(hub, connection, name, room)

	})

}
