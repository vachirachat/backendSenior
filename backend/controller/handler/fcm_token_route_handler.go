package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FCMRouteHandler struct {
	notifService *service.NotificationService
	authMw       *auth.JWTMiddleware
}

func NewFCMRouteHandler(notif *service.NotificationService, authMw *auth.JWTMiddleware) *FCMRouteHandler {
	return &FCMRouteHandler{
		notifService: notif,
		authMw:       authMw,
	}
}

func (handler *FCMRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.POST("/", handler.authMw.AuthRequired(), handler.handleRegsiterDevice)
	routerGroup.DELETE("/", handler.authMw.AuthRequired(), handler.handleUnregsiterDevice)
}

func (handler *FCMRouteHandler) handleRegsiterDevice(c *gin.Context) {
	userID := c.GetString(auth.UserIdField)

	var body struct {
		Token string `json:"token"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	err = handler.notifService.RegisterDevice(userID, body.Token)
	if err != nil {
		fmt.Println("[register device] error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *FCMRouteHandler) handleUnregsiterDevice(c *gin.Context) {
	userID := c.GetString(auth.UserIdField)

	var body struct {
		Token string `json:"token"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	tokens, err := handler.notifService.GetUserTokens(userID)
	if err != nil {
		fmt.Println("[delete device] check permission error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	found := false
	for _, tok := range tokens {
		if tok.Token == body.Token {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusForbidden, gin.H{"status": "you don't own the token"})
		return
	}

	err = handler.notifService.DeleteDevice(body.Token)
	if err != nil {
		fmt.Println("[delete device] error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
