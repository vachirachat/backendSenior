package routeAPI

import (
	"backendSenior/domain/service/auth"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

func AddAuthRoute(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	// //Test Google oAuth
	// route.LoadHTMLGlob("route/*")
	// route.GET("/", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "index.html", nil)
	// })

	// OauthGoogle
	routerGroup.GET("/auth/google/login", auth.OauthGoogleLogin)
	routerGroup.GET("/auth/google/callback", auth.OauthGoogleCallback)

}

func AddAuthRouteDev(routerGroup *gin.RouterGroup, connectionDB *mgo.Session) {

	// //Test Google oAuth
	// route.LoadHTMLGlob("route/*")
	// route.GET("/", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "index.html", nil)
	// })

	// OauthGoogle
	routerGroup.GET("/auth/google/login", auth.OauthGoogleLogin)
	routerGroup.GET("/auth/google/callback", auth.OauthGoogleCallback)

}
