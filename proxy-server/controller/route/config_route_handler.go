package route

import (
	g "common/utils/ginutils"
	"fmt"
	"log"
	"net/http"
	model_proxy "proxySenior/domain/model"
	"proxySenior/domain/service"
	"proxySenior/utils"

	"github.com/gin-gonic/gin"
)

// PingRouteHandler is a simple router that always reply with status OK
type ConfigRouteHandler struct {
	ConfigService *service.ConfigService
}

// NewPingRouteHandler create new ping route handler
func NewConfigRouteHandler(configService *service.ConfigService) *ConfigRouteHandler {
	return &ConfigRouteHandler{
		ConfigService: configService,
	}
}

// Mount add route to router group
func (handler *ConfigRouteHandler) Mount(routerGroup *gin.RouterGroup) {
<<<<<<< HEAD
	routerGroup.POST("docker/file", g.InjectGin(handler.configFileHandler))
	routerGroup.POST("docker/status", g.InjectGin(handler.configPluginNetworkStatus))
	routerGroup.GET("process/kill", g.InjectGin(handler.configKillProcess))
	routerGroup.GET("plugin/status", g.InjectGin(handler.configGetPluginStatus))
	routerGroup.GET("plugin/start", g.InjectGin(handler.proxySetPluginStart))
	routerGroup.GET("plugin/stop", g.InjectGin(handler.proxySetPluginStop))
=======
	routerGroup.POST("docker/code", handler.configCodeHandler)
	routerGroup.GET("docker/run/code", handler.configRunCodeProxy)
	routerGroup.POST("docker/file", handler.configFileHandler)
	routerGroup.POST("docker/status", handler.configPluginNetworkStatus)
	routerGroup.GET("process/kill", handler.configKillProcess)
	routerGroup.GET("plugin/status", handler.configGetPluginStatus)
	routerGroup.GET("plugin/start", handler.proxySetPluginStart)
	routerGroup.GET("plugin/stop", handler.proxySetPluginStop)
>>>>>>> feat/proxy/code-api

}

// Fix Just Debug
func (handler *ConfigRouteHandler) configGetPluginStatus(c *gin.Context, req struct{}) {
	c.JSON(http.StatusOK, gin.H{"status": handler.ConfigService.ConfigGetPluginStatus()})
}

// Fix Just Debug
func (handler *ConfigRouteHandler) proxySetPluginStart(c *gin.Context, req struct{}) {
	handler.ConfigService.ConfigSetStartProxy()
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (handler *ConfigRouteHandler) proxySetPluginStop(c *gin.Context, req struct{}) {
	handler.ConfigService.ConfigSetStopProxy()
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (handler *ConfigRouteHandler) configFileHandler(c *gin.Context, req struct{}) {
	c.Request.ParseMultipartForm(10 << 20)
	file, fileHandler, err := c.Request.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	err = handler.ConfigService.ConfigFileProxy(file, fileHandler)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err})
	}

	err = handler.ConfigService.ConfigStartPluginProcess("to_zip_" + utils.DOCKEREXEC_FILE_NAME)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err})
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (handler *ConfigRouteHandler) configKillProcess(c *gin.Context, req struct{}) {
	process_name, ok := c.Request.URL.Query()["process_name"]
	err := handler.ConfigService.ConfigStopPluginProcess(process_name[0])
	if err != nil || !ok {
		fmt.Println("Error configKillProcess the File")
		c.JSON(http.StatusInternalServerError, gin.H{"status": err})
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (handler *ConfigRouteHandler) configPluginNetworkStatus(c *gin.Context, req struct{}) {
	var storage model_proxy.JSONDocker
	err := c.ShouldBindJSON(&storage)
	if err != nil {
		log.Println("err -binding")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}
	resp, err := handler.ConfigService.ConfigPluginNetworkStatus(storage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err})
	}
	if resp == "" {
		c.JSON(http.StatusOK, gin.H{"status": "NO", "connect plugin with port": resp})
	}
	c.JSON(http.StatusOK, gin.H{"status": "OK", "connect plugin with port": resp})
}

func (handler *ConfigRouteHandler) configRunCodeProxy(c *gin.Context) {
	var storage model_proxy.JSONCODE
	err := c.ShouldBindJSON(&storage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})

	}
	err = handler.ConfigService.ConfigRunCodeProxy(storage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err})
	}
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (handler *ConfigRouteHandler) configCodeHandler(c *gin.Context) {
	var storage model_proxy.JSONCODE
	err := c.ShouldBindJSON(&storage)
	if err != nil {
		log.Println("err -binding")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}
	// Debug ->
	log.Println("c.Request.Body", storage)
	err = handler.ConfigService.ConfigCodeProxy(storage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err})
	}
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
