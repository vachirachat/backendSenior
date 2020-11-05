package route

import (
	"backendSenior/repository/pubsub"
	"backendSenior/route/routeAPI"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

func NewRouter(connectionDB *mgo.Session) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	routeAPI.AddUserRoute(api, connectionDB)
	routeAPI.AddRoomRoute(api, connectionDB)
	routeAPI.AddAuthRoute(api, connectionDB)
	routeAPI.AddMessageRoute(api, connectionDB)

	devAPI := router.Group("/dev")
	routeAPI.AddUserRouteDev(devAPI, connectionDB)
	routeAPI.AddRoomRouteDev(devAPI, connectionDB)
	routeAPI.AddAuthRouteDev(devAPI, connectionDB)
	routeAPI.AddMessageRouteDev(devAPI, connectionDB)

	//Test socket web
	router.LoadHTMLGlob("*.html")
	router.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "chat-room.html", nil)
	})

	router.GET("/connetSocket", func(context *gin.Context) {
		hub := pubsub.H
		go hub.Run()
		pubsub.ServeWs(context)
	})

	return router
}
