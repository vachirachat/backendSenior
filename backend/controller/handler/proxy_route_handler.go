package route

import (
	"backendSenior/controller/middleware/auth"
	authMw "backendSenior/controller/middleware/auth"
	"backendSenior/domain/dto"
	"backendSenior/domain/model"
	"backendSenior/domain/service"
	"backendSenior/utills"
	"bytes"
	g "common/utils/ginutils"
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
	routerGroup.GET("/", handler.authMw.AuthRequired("admin", "query"), g.InjectGin(handler.getAllProxies))
	routerGroup.POST("/", handler.authMw.AuthRequired("admin", "add"), g.InjectGin(handler.createProxy))
	routerGroup.POST("/:id/", handler.authMw.AuthRequired("admin", "add"), g.InjectGin(handler.updateProxy))
	routerGroup.GET("/:id/", handler.authMw.AuthRequired("admin", "view"), g.InjectGin(handler.getProxyByID))
	routerGroup.POST("/:id/reset", handler.authMw.AuthRequired("admin", "add"), g.InjectGin(handler.resetSecret))

	// Proxy-Org just debug
	routerGroup.GET("/:id/org", handler.authMw.AuthRequired("admin", "query"), g.InjectGin(handler.getOrgProxyByID))
	routerGroup.POST("/:id/org", handler.authMw.AuthRequired("admin", "add"), g.InjectGin(handler.addOrgProxyByID))
	routerGroup.DELETE("/:id/org", handler.authMw.AuthRequired("admin", "edit"), g.InjectGin(handler.removeOrgProxyByID))
	// routerGroup.GET("/:id/master-rooms", handler.getMasterRooms)

	// Proxy-Plugin
	routerGroup.POST("/:id/vm/file", handler.authMw.AuthRequired("admin", "add"), g.InjectGin(handler.forwardProxyVMFile))
	routerGroup.POST("/:id/vm/code", handler.authMw.AuthRequired("admin", "add"), g.InjectGin(handler.forwardProxyVMCode))
	routerGroup.POST("/:id/vm/code/kill", handler.authMw.AuthRequired("admin", "edit"), g.InjectGin(handler.forwardProxyVMKillCode))
	routerGroup.POST("/:id/vm/status", handler.authMw.AuthRequired("admin", "view"), g.InjectGin(handler.forwardProxyVMStatus))
	routerGroup.GET("/:id/process/:process_name/kill", handler.authMw.AuthRequired("admin", "edit"), g.InjectGin(handler.forwardProxyVMProcessKill))
	routerGroup.GET("/:id/plugin/status", handler.authMw.AuthRequired("admin", "edit"), g.InjectGin(handler.forwardProxyPluginStatus))
	routerGroup.GET("/:id/plugin/start", handler.authMw.AuthRequired("admin", "edit"), g.InjectGin(handler.forwardProxyPluginUP))
	routerGroup.GET("/:id/plugin/stop", handler.authMw.AuthRequired("admin", "edit"), g.InjectGin(handler.forwardProxyPluginDown))

}

func (handler *ProxyRouteHandler) getAllProxies(context *gin.Context, req struct{}) error {
	proxies, err := handler.proxyService.GetAll()
	if err != nil {
		log.Println("error roomListHandler", err.Error())
		return g.NewError(404, "bad Get listRoom")
	}
	context.JSON(http.StatusOK, proxies)
	return nil
}

// createProxy: new proxy with specified name, return ID and secret
func (handler *ProxyRouteHandler) createProxy(context *gin.Context, req struct{ Body dto.CreateProxyDto }) error {
	// var body model.Proxy

	// err := context.ShouldBindJSON(&body)
	// if err != nil || body.Name == "" {
	// 	fmt.Println(err)
	// 	context.JSON(http.StatusBadRequest, gin.H{"status": "bad data"})
	// 	return
	// }
	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	id, secret, err := handler.proxyService.NewProxy(b)
	if err != nil {
		return g.NewError(403, "bad Create Proxy ")
	}
	context.JSON(http.StatusOK, gin.H{"proxyId": id, "proxySecret": secret})
	return nil
}

// updateProxy: change name or IP or Port
func (handler *ProxyRouteHandler) updateProxy(context *gin.Context, req struct{ Body dto.UpdateProxyDto }) error {
	id := context.Param("id")
	if !bson.IsObjectIdHex(id) {
		// context.JSON(http.StatusBadRequest, gin.H{"status": "bad id"})
		return g.NewError(400, "bad proxy id ")
	}

	// var proxy model.Proxy
	// err := context.BindJSON(&proxy)
	// if !bson.IsObjectIdHex(id) {
	// 	context.JSON(http.StatusBadRequest, gin.H{"status": "bad proxy"})
	// 	return
	// }

	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	err = handler.proxyService.UpdateProxy(id, b.ToProxyUpdate())
	if err != nil {
		fmt.Println("error reset secret:", err.Error())
		// context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return g.NewError(430, "bad Update Proxy ")
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

func (handler *ProxyRouteHandler) getProxyByID(context *gin.Context, req struct{}) error {
	proxyID := context.Param("id")
	if !bson.IsObjectIdHex(proxyID) {
		return g.NewError(400, "bad proxy id ")
	}

	proxy, err := handler.proxyService.GetProxyByID(proxyID)
	if err != nil {
		return g.NewError(403, "bad get proxy")
	}

	context.JSON(http.StatusOK, proxy)
	return nil
}

func (handler *ProxyRouteHandler) resetSecret(context *gin.Context, req struct{}) error {
	id := context.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad proxy id ")
	}

	secret, err := handler.proxyService.ResetProxySecret(id)
	if err != nil {
		return g.NewError(403, "bad reset secret")
	}

	context.JSON(http.StatusOK, gin.H{"proxyId": id, "proxySecret": secret})
	return nil
}

// Debug data
func (handler *ProxyRouteHandler) getOrgProxyByID(context *gin.Context, req struct{}) error {
	id := context.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad proxy id ")
	}

	proxies, err := handler.proxyService.GetOrgProxyIDs(id)
	if err != nil {
		return g.NewError(404, "bad get proxy")
	}

	context.JSON(http.StatusOK, gin.H{"ordid": id, "proxies": proxies})
	return nil
}

func (handler *ProxyRouteHandler) addOrgProxyByID(context *gin.Context, req struct{ Body dto.UpdateProxyOrgDto }) error {
	id := context.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad proxy id ")
	}

	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	// var body model.Organize
	// err := context.ShouldBindJSON(&body)
	// if err != nil {
	// 	context.JSON(http.StatusBadRequest, gin.H{"status": err})
	// 	return
	// }

	err = handler.proxyService.AddProxiseToOrg(id, utills.ToStringArr(b.Proxies))
	if err != nil {
		return g.NewError(403, "bad add proxy to org")
	}

	context.JSON(http.StatusOK, gin.H{"ordid": id})
	return nil
}

func (handler *ProxyRouteHandler) removeOrgProxyByID(context *gin.Context, req struct{ Body dto.UpdateProxyOrgDto }) error {
	id := context.Param("id")
	if !bson.IsObjectIdHex(id) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad id"})
		return g.NewError(400, "bad proxy id ")
	}

	// var body model.Organize
	// err := context.ShouldBindJSON(&body)
	// if err != nil {
	// 	context.JSON(http.StatusBadRequest, gin.H{"status": err})
	// 	return
	// }

	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	err = handler.proxyService.RemoveProxiseFromOrg(id, utills.ToStringArr(b.Proxies))
	if err != nil {
		return g.NewError(404, "bad delete proxy to org")
	}

	context.JSON(http.StatusOK, gin.H{"ordid": id})
	return nil
}

// Debug data

// Fix - forward to more resource full
// Forward API to porxy by proxyID query IP -> "docker/file"
func (handler *ProxyRouteHandler) forwardProxyVMFile(context *gin.Context, req struct{}) error {
	reps, err := handler.forwardAPI("docker/file", context)
	if err != nil {
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return nil
}

// Forward API to porxy by proxyID query IP -> "docker/code"
func (handler *ProxyRouteHandler) forwardProxyVMCode(context *gin.Context, req struct{}) error {
	reps, err := handler.forwardAPI("docker/code", context)
	if err != nil {
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return nil
}

// Forward API to porxy by proxyID query IP -> "docker/code/kill"
func (handler *ProxyRouteHandler) forwardProxyVMKillCode(context *gin.Context, req struct{}) error {
	reps, err := handler.forwardAPI("docker/code/kill", context)
	if err != nil {
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return nil
}

// Forward API to porxy by proxyID query IP -> "docker/status"
func (handler *ProxyRouteHandler) forwardProxyVMStatus(context *gin.Context, req struct{}) error {
	reps, err := handler.forwardAPI("docker/status", context)
	if err != nil {
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return nil
}

// Forward API to porxy by proxyID query IP -> "process/kill?process_name="+proxyProcess
func (handler *ProxyRouteHandler) forwardProxyVMProcessKill(context *gin.Context, req struct{}) error {
	proxyProcess := context.Param("process_name")
	reps, err := handler.forwardAPI("process/kill?process_name="+proxyProcess, context)
	if err != nil {
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return nil
}

// Forward API to porxy by proxyID query IP ->  "plugin/status"
func (handler *ProxyRouteHandler) forwardProxyPluginStatus(context *gin.Context, req struct{}) error {
	reps, err := handler.forwardAPI("plugin/status", context)
	if err != nil {
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return nil
}

// Forward API to porxy by proxyID query IP -> "plugin/start"
func (handler *ProxyRouteHandler) forwardProxyPluginUP(context *gin.Context, req struct{}) error {
	reps, err := handler.forwardAPI("plugin/start", context)
	if err != nil {
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return nil
}

// Forward API to porxy by proxyID query IP -> "plugin/stop"
func (handler *ProxyRouteHandler) forwardProxyPluginDown(context *gin.Context, req struct{}) error {
	reps, err := handler.forwardAPI("plugin/stop", context)
	if err != nil {
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": reps})
	return nil

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
