package route

import (
	"backendSenior/controller/middleware/auth"
	authMw "backendSenior/controller/middleware/auth"
	"backendSenior/domain/model"
	"backendSenior/domain/service"
	"backendSenior/utills"
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
	roomService  *service.RoomService // get master proxy
	userService  *service.UserService
	authMw       *auth.JWTMiddleware
	validate     *utills.StructValidator
}

// NewProxyRouteHandler create new handler for proxy
func NewProxyRouteHandler(proxyService *service.ProxyService, roomService *service.RoomService, userService *service.UserService, authMw *auth.JWTMiddleware, validate *utills.StructValidator) *ProxyRouteHandler {
	return &ProxyRouteHandler{
		proxyService: proxyService,
		roomService:  roomService,
		authMw:       authMw,
		validate:     validate,
		userService:  userService,
	}
}

//Mount make RoomRouteHandler handler request from specific `RouterGroup`
func (handler *ProxyRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	// Proxy
	routerGroup.GET("/", handler.getAllProxies)
	routerGroup.POST("/", handler.createProxy)
	routerGroup.POST("/:id/", handler.updateProxy)
	routerGroup.GET("/:id/", handler.getProxyByID)
	routerGroup.POST("/:id/reset", handler.resetSecret)

	// Proxy-Org just debug
	routerGroup.GET("/:id/org", handler.getOrgProxyByID)
	routerGroup.POST("/:id/org", handler.addOrgProxyByID)
	routerGroup.DELETE("/:id/org", handler.removeOrgProxyByID)
	// routerGroup.GET("/:id/master-rooms", handler.getMasterRooms)

	// Proxy-Plugin
	routerGroup.POST("/:id/vm/file", handler.authMw.AuthRequired("admin", "query"), handler.forwardProxyVMFile)
	routerGroup.POST("/:id/vm/code", handler.authMw.AuthRequired("admin", "query"), handler.forwardProxyVMCode)
	routerGroup.POST("/:id/vm/code/kill", handler.authMw.AuthRequired("admin", "query"), handler.forwardProxyVMKillCode)
	routerGroup.POST("/:id/vm/status", handler.authMw.AuthRequired("admin", "query"), handler.forwardProxyVMStatus)
	routerGroup.GET("/:id/process/:process_name/kill", handler.authMw.AuthRequired("admin", "query"), handler.forwardProxyVMProcessKill)
	routerGroup.GET("/:id/plugin/status", handler.authMw.AuthRequired("admin", "query"), handler.forwardProxyPluginStatus)
	routerGroup.GET("/:id/plugin/start", handler.authMw.AuthRequired("admin", "query"), handler.forwardProxyPluginUP)
	routerGroup.GET("/:id/plugin/stop", handler.authMw.AuthRequired("admin", "query"), handler.forwardProxyPluginDown)

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

func (handler *ProxyRouteHandler) getProxyByID(context *gin.Context) {
	proxyID := context.Param("id")
	if !bson.IsObjectIdHex(proxyID) || proxyID == "" {
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

// Debug data
func (handler *ProxyRouteHandler) getOrgProxyByID(context *gin.Context) {
	id := context.Param("id")
	if !bson.IsObjectIdHex(id) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad id"})
		return
	}

	proxies, err := handler.proxyService.GetOrgProxyIDs(id)
	if err != nil {
		fmt.Println("error reset secret:", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"ordid": id, "proxies": proxies})
}

func (handler *ProxyRouteHandler) addOrgProxyByID(context *gin.Context) {
	id := context.Param("id")
	if !bson.IsObjectIdHex(id) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad id"})
		return
	}

	var body model.Organize
	err := context.ShouldBindJSON(&body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": err})
		return
	}

	err = handler.proxyService.AddProxiseToOrg(id, utills.ToStringArr(body.Proxies))
	if err != nil {
		fmt.Println("error reset secret:", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"ordid": id})
}

func (handler *ProxyRouteHandler) removeOrgProxyByID(context *gin.Context) {
	id := context.Param("id")
	if !bson.IsObjectIdHex(id) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad id"})
		return
	}

	var body model.Organize
	err := context.ShouldBindJSON(&body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": err})
		return
	}

	err = handler.proxyService.RemoveProxiseFromOrg(id, utills.ToStringArr(body.Proxies))
	if err != nil {
		fmt.Println("error reset secret:", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"ordid": id})
}

// Debug data

// Fix - forward to more resource full
// Forward API to porxy by proxyID query IP -> "docker/file"
func (handler *ProxyRouteHandler) forwardProxyVMFile(context *gin.Context) {
	reps, err := handler.forwardAPI("docker/file", context)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return
}

// Forward API to porxy by proxyID query IP -> "docker/code"
func (handler *ProxyRouteHandler) forwardProxyVMCode(context *gin.Context) {
	reps, err := handler.forwardAPI("docker/code", context)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return
}

// Forward API to porxy by proxyID query IP -> "docker/code/kill"
func (handler *ProxyRouteHandler) forwardProxyVMKillCode(context *gin.Context) {
	reps, err := handler.forwardAPI("docker/code/kill", context)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return
}

// Forward API to porxy by proxyID query IP -> "docker/status"
func (handler *ProxyRouteHandler) forwardProxyVMStatus(context *gin.Context) {
	reps, err := handler.forwardAPI("docker/status", context)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return
}

// Forward API to porxy by proxyID query IP -> "process/kill?process_name="+proxyProcess
func (handler *ProxyRouteHandler) forwardProxyVMProcessKill(context *gin.Context) {
	proxyProcess := context.Param("process_name")
	reps, err := handler.forwardAPI("process/kill?process_name="+proxyProcess, context)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return
}

// Forward API to porxy by proxyID query IP ->  "plugin/status"
func (handler *ProxyRouteHandler) forwardProxyPluginStatus(context *gin.Context) {
	reps, err := handler.forwardAPI("plugin/status", context)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return
}

// Forward API to porxy by proxyID query IP -> "plugin/start"
func (handler *ProxyRouteHandler) forwardProxyPluginUP(context *gin.Context) {
	reps, err := handler.forwardAPI("plugin/start", context)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return
}

// Forward API to porxy by proxyID query IP -> "plugin/stop"
func (handler *ProxyRouteHandler) forwardProxyPluginDown(context *gin.Context) {
	reps, err := handler.forwardAPI("plugin/stop", context)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return

}

// Fix Refactor : More reuse code
func (handler *ProxyRouteHandler) forwardAPI(path string, context *gin.Context) (string, error) {
	var reps string
	isFull := context.Query("tagID") == "ok"
	log.Println("context  Query>>", path, context.Query("tagID"))
	if isFull {
		id := context.GetString(authMw.UserIdField)
		user, err := handler.userService.GetUserByID(id)
		if err != nil {
			return reps, err
		}
		for _, orgv := range user.Organize {
			proxise, _ := handler.proxyService.GetOrgProxyIDs(orgv.Hex())
			for _, v := range proxise {
				reps, err := forwardToProxy(v, path, context)
				if err != nil {
					return reps, err
				}
			}
		}

	} else {
		proxy := handler.templateInput(context)
		reps, err := forwardToProxy(proxy, path, context)
		if err != nil {
			return reps, err
		}
	}
	return reps, nil
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
func (handler *ProxyRouteHandler) templateInput(context *gin.Context) model.Proxy {
	proxyID := context.Param("id")
	if !bson.IsObjectIdHex(proxyID) || proxyID == "" {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad proxy ID"})
		return model.Proxy{}
	}
	proxy, err := handler.proxyService.GetProxyByID(proxyID)
	if err != nil {
		fmt.Println("error get proxy by id", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return model.Proxy{}
	}
	return proxy
}
