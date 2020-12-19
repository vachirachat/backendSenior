package route

import (
	"backendSenior/domain/service"
	"fmt"

	"backendSenior/domain/model"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
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
	routerGroup.POST("/:id/reset", handler.resetSecret)
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

// createProxy: new proxy with specified name, return ID and secret
func (handler *ProxyRouteHandler) createProxy(context *gin.Context) {
	var body model.Proxy

	err := context.ShouldBindJSON(&body)
	if err != nil || body.Name == "" {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad data"})
		return
	}

	id, secret, err := handler.proxyService.NewProxy(body.Name)
	if err != nil {
		log.Println("error createProxy", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"proxyId": id, "proxySecret": secret})
}

func (handler *ProxyRouteHandler) resetSecret(context *gin.Context) {
	id := context.Param("id")
	if !bson.IsObjectIdHex(id) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad id"})
		return
	}

	secret, err := handler.proxyService.ResetProxySecret(id)
	if err != nil {
		fmt.Println("error reset secret:", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"proxyId": id, "proxySecret": secret})
}
