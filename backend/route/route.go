package route

import (
	"backendSenior/route/routeAPI"

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
	return router
}
