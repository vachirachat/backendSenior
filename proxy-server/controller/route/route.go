package route

import (
	"proxySenior/domain/service"
	"proxySenior/domain/service/key_service"
	"runtime"
	"time"

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
	KeyService        *key_service.KeyService
	FileService       *service.FileService
	Encrpytion        *service.EncryptionService
}

// NewRouter create router from deps
func (deps *RouterDeps) NewRouter() *gin.Engine {
	authMiddleware := middleware.NewAuthMiddleware(deps.AuthService)
	r := gin.Default()

	chatRouteHandler := NewChatRouteHandler(deps.UpstreamService, deps.DownstreamService, authMiddleware, deps.Encrpytion)
	messageRouteHandler := NewMessageRouteHandler(deps.MessageService, deps.KeyService)
	pingRouteHandler := NewPingRouteHandler()
	configRouteHandler := NewConfigRouteHandler(deps.ConfigService)
	keyRouteHandler := NewKeyRouteHandler(deps.KeyService)
	fileRouteHandler := NewFileRouteHandler(deps.FileService, authMiddleware, deps.KeyService, deps.UpstreamService)

	pingRouteHandler.Mount(r.Group("/ping"))

	subgroup := r.Group("/api/v1")
	configRouteHandler.Mount(subgroup.Group("/config"))
	chatRouteHandler.Mount(subgroup.Group("/chat"))
	messageRouteHandler.Mount(subgroup.Group("/message"))
	keyRouteHandler.Mount(subgroup.Group("/key"))
	fileRouteHandler.Mount(subgroup.Group("/file"))

	r.GET("/debug/gc", func(c *gin.Context) {
		t1 := time.Now()
		runtime.GC()
		t2 := time.Now()
		c.JSON(200, gin.H{"took": t2.Sub(t1).Milliseconds()})
	})

	return r
}
