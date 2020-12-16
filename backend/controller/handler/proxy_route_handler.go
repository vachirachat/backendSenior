package route

import (
	"backendSenior/domain/service"

	"backendSenior/domain/model"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProxyRouteHandler struct {
	proxyService *service.ProxyService
}

// NewProxyRouteHandler create new handler for proxy
func NewProxyRouteHandler(proxyService *service.ProxyService) *ProxyRouteHandler {
	return &ProxyRouteHandler{
		proxyService: proxyService,
	}
}

//Mount make RoomRouteHandler handler request from specific `RouterGroup`
func (handler *ProxyRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("/", handler.getAllProxies)
	routerGroup.POST("/", handler.createProxy)
}

func (handler *ProxyRouteHandler) getAllProxies(context *gin.Context) {
	proxies, err := handler.proxyService.GetAll()
	if err != nil {
		log.Println("error roomListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, proxies)
}

// for get room by id
func (handler *ProxyRouteHandler) createProxy(context *gin.Context) {
	var body model.Proxy

	err := context.ShouldBindJSON(&body)
	if err != nil || body.Name == "" {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad data"})
	}

	id, err := handler.proxyService.NewProxy(body.Name)
	if err != nil {
		log.Println("error createProxy", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"proxyID": id})
}
