package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PingRouteHandler is a simple router that always reply with status OK
type PingRouteHandler struct{}

// NewPingRouteHandler create new ping route handler
func NewPingRouteHandler() *PingRouteHandler {
	return &PingRouteHandler{}
}

// Mount add route to router group
func (handler *PingRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("", handler.pongHandler)
}

func (handler *PingRouteHandler) pongHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
