package route

import (
	authMw "backendSenior/controller/middleware/auth"
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"

	"github.com/gin-gonic/gin"
)

// RouterDeps declare dependency for router, it's used to create route handlers
type RouterDeps struct {
	MessageService   *service.MessageService
	RoomService      *service.RoomService
	UserService      *service.UserService
	ChatService      *service.ChatService
	ProxyService     *service.ProxyService
	JWTService       *auth.JWTService
	ProxyAuth        *auth.ProxyAuth
	OraganizeService *service.OrganizeService
}

// NewRouter create new router (gin server) with various handler
func (deps *RouterDeps) NewRouter() *gin.Engine {
	// create middleware first
	authMiddleware := authMw.NewJWTMiddleware(deps.JWTService)
	proxyMw := authMw.NewProxyMiddleware(deps.ProxyAuth)

	// create handler (some require middleware)
	roomRouteHandler := NewRoomRouteHandler(deps.RoomService, authMiddleware)
	userRouteHandler := NewUserRouteHandler(deps.UserService, deps.JWTService, authMiddleware)
	messageRouteHandler := NewMessageRouteHandler(deps.MessageService)
	chatRouteHandler := NewChatRouteHandler(deps.ChatService, proxyMw, deps.RoomService)
	proxyRouteHandler := NewProxyRouteHandler(deps.ProxyService)
	OrganizeRouteHandler := NewOrganizeRouteHandler(deps.OraganizeService, deps.JWTService)
	r := gin.Default()

	subgroup := r.Group("/api/v1")

	roomRouteHandler.Mount(subgroup.Group("/room"))
	userRouteHandler.Mount(subgroup.Group("/user"))
	messageRouteHandler.Mount(subgroup.Group("/message"))
	chatRouteHandler.Mount(subgroup.Group("/chat"))
	proxyRouteHandler.Mount(subgroup.Group("/proxy"))
	OrganizeRouteHandler.Mount(subgroup.Group("/org"))
	return r
}
