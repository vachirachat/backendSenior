package route

import (
	"backendSenior/domain/model"
	"backendSenior/domain/service"
	"bytes"
	"fmt"
	"io/ioutil"
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
	routerGroup.POST("/:id/", handler.updateProxy)
	routerGroup.GET("/:id/", handler.getProxyByID)
	routerGroup.POST("/:id/reset", handler.resetSecret)
	routerGroup.POST("/:id/file", handler.forwardProxyFile)
	routerGroup.POST("/:id/status", handler.forwardProxyStatus)
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

func (handler *ProxyRouteHandler) getProxyByID(context *gin.Context) {
	proxyID := context.Param("id")
	if !bson.IsObjectIdHex(proxyID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad proxy ID"})
		return
	}

	proxy, err := handler.proxyService.GetProxyByID(proxyID)
	if err != nil {
		fmt.Println("error get proxy by id", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
	}

	context.JSON(http.StatusOK, proxy)
}

// createProxy: new proxy with specified name, return ID and secret
func (handler *ProxyRouteHandler) createProxy(context *gin.Context) {
	var body model.Proxy

	err := context.ShouldBindJSON(&body)
	if err != nil || body.Name == "" {
		fmt.Println(err)
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad data"})
		return
	}

	id, secret, err := handler.proxyService.NewProxy(body)
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

// updateProxy: change name or IP
func (handler *ProxyRouteHandler) updateProxy(context *gin.Context) {
	id := context.Param("id")
	if !bson.IsObjectIdHex(id) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad id"})
		return
	}

	var proxy model.Proxy
	err := context.BindJSON(&proxy)
	if !bson.IsObjectIdHex(id) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad proxy"})
		return
	}

	err = handler.proxyService.UpdateProxy(id, proxy)
	if err != nil {
		fmt.Println("error reset secret:", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *ProxyRouteHandler) forwardProxyFile(context *gin.Context) {
	proxyID := context.Param("id")
	if !bson.IsObjectIdHex(proxyID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad proxy ID"})
		return
	}
	proxy, err := handler.proxyService.GetProxyByID(proxyID)
	if err != nil {
		fmt.Println("error get proxy by id", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
	}
	resp, err := forwardToProxy(proxy, "docker/file", context)
	if err != nil {
		fmt.Println("error get proxy by id", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
	}
	context.JSON(http.StatusOK, gin.H{"status": resp})
}

func (handler *ProxyRouteHandler) forwardProxyStatus(context *gin.Context) {
	proxyID := context.Param("id")
	if !bson.IsObjectIdHex(proxyID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad proxy ID"})
		return
	}
	proxy, err := handler.proxyService.GetProxyByID(proxyID)
	if err != nil {
		fmt.Println("error get proxy by id", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
	}
	resp, err := forwardToProxy(proxy, "docker/status", context)
	if err != nil {
		fmt.Println("error get proxy by id", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
	}
	context.JSON(http.StatusOK, gin.H{"status": resp})
}

func (handler *ProxyRouteHandler) forwardProxyDown(context *gin.Context) {
	proxyID := context.Param("id")
	if !bson.IsObjectIdHex(proxyID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad proxy ID"})
		return
	}
	proxy, err := handler.proxyService.GetProxyByID(proxyID)
	if err != nil {
		fmt.Println("error get proxy by id", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
	}
	resp, err := forwardToProxy(proxy, "docker/status", context)
	if err != nil {
		fmt.Println("error get proxy by id", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
	}
	context.JSON(http.StatusOK, gin.H{"status": resp})
}

// Fix Refactor : Move to Service
func forwardToProxy(proxy model.Proxy, path string, context *gin.Context) (string, error) {
	// we need to buffer the body if we want to read it here and send it
	// in the request.
	body, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		return "", err
	}

	// you can reassign the body if you need to parse it as multipart
	context.Request.Body = ioutil.NopCloser(bytes.NewReader(body))

	proxyScheme := "http"
	proxyHost := proxy.IP + ":" + fmt.Sprint(proxy.Port)
	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("%s://%s%s", proxyScheme, proxyHost, "/api/v1/config/"+path)
	log.Println(url)
	proxyReq, err := http.NewRequest(context.Request.Method, url, bytes.NewReader(body))
	proxyReq.Header = make(http.Header)
	for h, val := range context.Request.Header {
		proxyReq.Header[h] = val
	}
	httpClient := http.Client{}
	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	return string(body), nil
}

// Fix Refactor : Move to Service
