package route

import (
	"backendSenior/api"
	"backendSenior/route/routeAPI"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

func setStaticFolder(route *gin.Engine) {

	route.StaticFS("./public", http.Dir("./public/"))
}

func NewRouter(connectionDB *mgo.Session) *gin.Engine {
	router := gin.Default()
	apiRoute := router.Group("/api")
	routeAPI.AddUserRoute(apiRoute, connectionDB)
	routeAPI.AddRoomRoute(apiRoute, connectionDB)
	routeAPI.AddAuthRoute(apiRoute, connectionDB)
	routeAPI.AddMessageRoute(apiRoute, connectionDB)

	devAPI := router.Group("/dev")
	routeAPI.AddUserRouteDev(devAPI, connectionDB)
	routeAPI.AddRoomRouteDev(devAPI, connectionDB)
	routeAPI.AddAuthRouteDev(devAPI, connectionDB)
	routeAPI.AddMessageRouteDev(devAPI, connectionDB)

	setStaticFolder(router)

	router.GET("/", api.RenderHome)

	router.GET("/ws", api.SocketConnect)

	return router
}
