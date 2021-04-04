package route

import (
	"proxySenior/domain/service"

	"proxySenior/controller/middleware"

	"github.com/gin-gonic/gin"
)

// RouterDeps declare dependency for controller
type RouterDeps struct {
	UpstreamService   *service.ChatUpstreamService
	DownstreamService *service.ChatDownstreamService
	AuthService       *service.DelegateAuthService
	MessageService    *service.MessageService
	ConfigService     *service.ConfigService
}

// NewRouter create router from deps
func (deps *RouterDeps) NewRouter() *gin.Engine {
	authMiddleware := middleware.NewAuthMiddleware(deps.AuthService)

	r := gin.Default()

	chatRouteHandler := NewChatRouteHandler(deps.UpstreamService, deps.DownstreamService, authMiddleware)
	messageRouteHandler := NewMessageRouteHandler(deps.MessageService)
	pingRouteHandler := NewPingRouteHandler()
	configRouteHandler := NewConfigRouteHandler(deps.ConfigService)
	pingRouteHandler.Mount(r.Group("/ping"))

	subgroup := r.Group("/api/v1")
	configRouteHandler.Mount(subgroup.Group("/config"))
	chatRouteHandler.Mount(subgroup.Group("/chat"))
	messageRouteHandler.Mount(subgroup.Group("/message"))

	return r
}
