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
}

// NewRouter create router from deps
func (deps *RouterDeps) NewRouter() *gin.Engine {
	authMiddleware := middleware.NewAuthMiddleware(deps.AuthService)

	r := gin.Default()

	chatRouteHandler := NewChatRouteHandler(deps.UpstreamService, deps.DownstreamService, authMiddleware)
	messageRouteHandler := NewMessageRouteHandler(deps.MessageService)

	subgroup := r.Group("/api/v1")

	chatRouteHandler.Mount(subgroup.Group("/chat"))
	messageRouteHandler.Mount(subgroup.Group("/message"))

	return r
}
