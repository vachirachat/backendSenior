package route

import (
	"proxySenior/domain/service"
	"proxySenior/domain/service/key_service"

	"proxySenior/controller/middleware"

	"github.com/gin-gonic/gin"
)

// RouterDeps declare dependency for controller
type RouterDeps struct {
	UpstreamService   *service.ChatUpstreamService
	DownstreamService *service.ChatDownstreamService
	AuthService       *service.DelegateAuthService
	MessageService    *service.MessageService
	KeyService        *key_service.KeyService
}

// NewRouter create router from deps
func (deps *RouterDeps) NewRouter() *gin.Engine {
	authMiddleware := middleware.NewAuthMiddleware(deps.AuthService)
	r := gin.Default()

	chatRouteHandler := NewChatRouteHandler(deps.UpstreamService, deps.DownstreamService, authMiddleware, deps.KeyService)
	messageRouteHandler := NewMessageRouteHandler(deps.MessageService)
	pingRouteHandler := NewPingRouteHandler()
	keyRouteHandler := NewKeyRouteHandler(deps.KeyService)

	pingRouteHandler.Mount(r.Group("/ping"))

	subgroup := r.Group("/api/v1")

	chatRouteHandler.Mount(subgroup.Group("/chat"))
	messageRouteHandler.Mount(subgroup.Group("/message"))
	keyRouteHandler.Mount(subgroup.Group("/key"))

	return r
}
