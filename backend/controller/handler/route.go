package route

import (
	authMw "backendSenior/controller/middleware/auth"
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"
	"backendSenior/utills"

	"github.com/gin-gonic/gin"
)

// RouterDeps declare dependency for router, it's used to create route handlers
type RouterDeps struct {
	MessageService      *service.MessageService
	RoomService         *service.RoomService
	UserService         *service.UserService
	ChatService         *service.ChatService
	ProxyService        *service.ProxyService
	JWTService          *auth.JWTService
	ProxyAuth           *auth.ProxyAuth
	OraganizeService    *service.OrganizeService
	NotificationService *service.NotificationService
	KeyExchangeService  *service.KeyExchangeService
	FileService         *service.FileService
	StickerService      *service.StickerService
	Validate            *utills.StructValidator
}

// NewRouter create new router (gin server) with various handler
func (deps *RouterDeps) NewRouter() *gin.Engine {
	// create middleware first
	authMiddleware := authMw.NewJWTMiddleware(deps.JWTService, deps.RoomService, deps.OraganizeService)
	proxyMw := authMw.NewProxyMiddleware(deps.ProxyAuth)

	// create handler (some require middleware)
	roomRouteHandler := NewRoomRouteHandler(deps.RoomService, authMiddleware, deps.UserService, deps.ProxyService, deps.ChatService, deps.OraganizeService, deps.KeyExchangeService, deps.FileService, deps.Validate)
	userRouteHandler := NewUserRouteHandler(deps.UserService, deps.JWTService, authMiddleware, deps.FileService)
	messageRouteHandler := NewMessageRouteHandler(deps.MessageService, deps.FileService, deps.RoomService, authMiddleware, deps.Validate)
	chatRouteHandler := NewChatRouteHandler(deps.ChatService, proxyMw, deps.RoomService, deps.KeyExchangeService, deps.Validate)
	proxyRouteHandler := NewProxyRouteHandler(deps.ProxyService, deps.RoomService, deps.UserService, authMiddleware, deps.Validate)
	organizeRouteHandler := NewOrganizeRouteHandler(deps.OraganizeService, authMiddleware, deps.UserService, deps.ProxyService, deps.RoomService, deps.Validate)
	fcmTokenRouteHandler := NewFCMRouteHandler(deps.NotificationService, authMiddleware, deps.Validate)
	connStateRouteHandler := NewConnStateRouteHandler(deps.NotificationService, authMiddleware, deps.Validate)
	keyRouteHandler := NewKeyRoute(deps.ProxyService, deps.KeyExchangeService, deps.ChatService, deps.Validate)
	fileRouteHandler := NewFileRouteHandler(deps.FileService, deps.RoomService, authMiddleware, deps.Validate)
	StickerRouteHandler := NewStickerRouteHandler(deps.StickerService, authMiddleware, deps.Validate)
	r := gin.New()
	r.Use(gin.Recovery())

	subgroup := r.Group("/api/v1")

	roomRouteHandler.Mount(subgroup.Group("/room"))
	userRouteHandler.Mount(subgroup.Group("/user"))
	messageRouteHandler.Mount(subgroup.Group("/message"))
	chatRouteHandler.Mount(subgroup.Group("/chat"))
	proxyRouteHandler.Mount(subgroup.Group("/proxy"))
	organizeRouteHandler.Mount(subgroup.Group("/org"))
	fcmTokenRouteHandler.Mount(subgroup.Group("/fcm"))
	connStateRouteHandler.Mount(subgroup.Group("/conn"))
	keyRouteHandler.Mount(subgroup.Group("/key"))
	fileRouteHandler.Mount(subgroup.Group("/file"))
	StickerRouteHandler.Mount(subgroup.Group("/sticker"))

	v2 := r.Group("/api/v2")
	organizeRouteHandler.MountV2(v2.Group("/org"))

	return r
}
