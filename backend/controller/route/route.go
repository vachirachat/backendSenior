package route

import (
	"backendSenior/controller/route/routeAPI"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

func NewRouter(connectionDB *mgo.Session) *gin.Engine {
	router := gin.Default()
	apiRoute := router.Group("/api")
	
	routeAPI.AddUserRoute(apiRoute, connectionDB)
	routeAPI.AddRoomRoute(apiRoute, connectionDB)
	routeAPI.AddAuthRoute(apiRoute, connectionDB)
	
	routeAPI.AddMessageRoute(apiRoute, connectionDB)



	devAPI := router.Group("/dev")
	// routeAPI.AddUserRouteDev(devAPI, connectionDB)
	// routeAPI.AddRoomRouteDev(devAPI, connectionDB)
	routeAPI.AddAuthRouteDev(devAPI, connectionDB)
	// routeAPI.AddMessageRouteDev(devAPI, connectionDB)

	socketRoute := router.Group("/")
	routeAPI.AddCoversationRoute(socketRoute, connectionDB)
	return router
}
