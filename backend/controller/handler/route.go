package route

import (
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"

	"github.com/gin-gonic/gin"
)

// RouterDeps declare dependency for router, it's used to create route handlers
type RouterDeps struct {
	MessageService *service.MessageService
	RoomService    *service.RoomService
	UserService    *service.UserService
	AuthService    *auth.AuthService
	ChatService    *service.ChatService
}

// NewRouter create new router (gin server) with various handler
func (deps *RouterDeps) NewRouter() *gin.Engine {

	roomRouteHandler := NewRoomRouteHandler(deps.RoomService, deps.AuthService)
	userRouteHandler := NewUserRouteHandler(deps.UserService)
	messageRouteHandler := NewMessageRouteHandler(deps.MessageService, deps.AuthService)
	chatRouteHandler := NewChatRouteHandler(deps.ChatService)

	r := gin.Default()

	subgroup := r.Group("/api/v1")

	roomRouteHandler.Mount(subgroup.Group("/room"))
	userRouteHandler.Mount(subgroup) // this subroute isn't restful so I mount like this
	messageRouteHandler.Mount(subgroup.Group("/message"))
	chatRouteHandler.Mount(subgroup.Group("/chat"))

	return r
}
