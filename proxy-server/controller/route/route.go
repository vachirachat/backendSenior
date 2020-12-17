package route

import (
	"proxySenior/domain/service"

	"github.com/gin-gonic/gin"
)

// RouterDeps declare dependency for controller
type RouterDeps struct {
	UpstreamService   *service.ChatUpstreamService
	DownstreamService *service.ChatDownstreamService
}

// NewRouter create router from deps
func (deps *RouterDeps) NewRouter() *gin.Engine {

	r := gin.Default()

	chatRouteHandler := NewChatRouteHandler(deps.UpstreamService, deps.DownstreamService)

	subgroup := r.Group("/api/v1")

	chatRouteHandler.Mount(subgroup.Group("/chat"))

	return r
}
