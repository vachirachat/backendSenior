package api

import (
	socket "backendSenior/repository/pubsub-chat"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// RenderHome Rendering the Home Page
func RenderHome(context *gin.Context) {
	http.ServeFile(context.Writer, context.Request, "views/index.html")
}

func SocketConnect(context *gin.Context) {
	w := context.Writer
	r := context.Request

	hub := socket.NewHub()
	go hub.Run()

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

	// Query Room From Database -< check to query in API or in Socket
	room := &socket.Room{}

	socket.CreateNewSocketUser(hub, connection, name, room)

}
