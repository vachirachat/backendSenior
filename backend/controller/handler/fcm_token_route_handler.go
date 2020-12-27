package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/model"
	"backendSenior/domain/service"
	"fmt"
	"net/http"
	"time"

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
	routerGroup.GET("/", handler.authMw.AuthRequired(), handler.handleGetUserDevices)
	routerGroup.DELETE("/", handler.authMw.AuthRequired(), handler.handleUnregsiterDevice)
	routerGroup.POST("/ping", handler.authMw.AuthRequired(), handler.handlePing)
	routerGroup.POST("/check-status", handler.authMw.AuthRequired(), handler.checkTokenStatus)
	routerGroup.POST("/test-notification", handler.authMw.AuthRequired(), handler.sendTestNotification)
}

func (handler *FCMRouteHandler) handleRegsiterDevice(c *gin.Context) {
	userID := c.GetString(auth.UserIdField)

	var body model.FCMToken
	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	tok, err := handler.notifService.GetTokenByID(body.Token)

	// found
	if tok.Token != "" {
		if tok.UserID.Hex() != userID {
			c.JSON(http.StatusForbidden, gin.H{"status": "token already used by another user"})
			return
		}

		err = handler.notifService.RefreshDevice(tok.Token)
		if err != nil {
			fmt.Println("[register device] refresh error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "refreshed"})
		return
	}

	// not found
	err = handler.notifService.RegisterDevice(userID, body.Token, body.DeviceName)
	if err != nil {
		fmt.Println("[register device] error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *FCMRouteHandler) handleGetUserDevices(c *gin.Context) {
	userID := c.GetString(auth.UserIdField)

	tokens, err := handler.notifService.GetUserTokens(userID)

	if err != nil {
		if err.Error() == "not found" {
			c.JSON(http.StatusOK, []model.FCMToken{})
			return
		} else {
			fmt.Println("[get user device] error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
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

// return status; <not found>, <owned>, <owned by other>
func (handler *FCMRouteHandler) checkTokenStatus(c *gin.Context) {
	var body model.FCMToken

	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	userID := c.GetString(auth.UserIdField)

	token, err := handler.notifService.GetTokenByID(body.Token)

	if err != nil {
		if err.Error() == "not found" {
			c.JSON(http.StatusOK, gin.H{"status": "not found"})
			return
		}
		fmt.Println("[check token status] error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
	}

	if token.UserID.Hex() == userID {
		c.JSON(http.StatusOK, gin.H{"status": "owned"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "owned by other"})
	}
}

func (handler *FCMRouteHandler) handlePing(c *gin.Context) {

	var body struct {
		Token string `json:"token"`
	}

	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad token"})
		return
	}

	userID := c.GetString(auth.UserIdField)
	tok, err := handler.notifService.GetTokenByID(body.Token)
	if err != nil {
		fmt.Println("[notification ping] error", err)
		c.JSON(http.StatusForbidden, gin.H{"status": "something went wrong"})
		return
	} else if tok.UserID.Hex() != userID {
		c.JSON(http.StatusForbidden, gin.H{"status": "not you token"})
		return
	}

	err = handler.notifService.SetLastSeenTime(body.Token, time.Now())
	if err != nil {
		fmt.Println("[notification ping] update last seen", err)
		c.JSON(http.StatusForbidden, gin.H{"status": "something went wrong"})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (handler *FCMRouteHandler) sendTestNotification(c *gin.Context) {
	var body model.FCMToken

	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	tok, err := handler.notifService.GetTokenByID(body.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "something went wrong, ensure that token exists",
		})
		return
	}

	userID := c.GetString(auth.UserIdField)

	if tok.UserID.Hex() != userID {
		c.JSON(http.StatusForbidden, gin.H{"status": "not your own token"})
		return
	}

	sent, err := handler.notifService.SendNotifications([]string{body.Token}, &model.Notification{
		Title: "Test notification",
		Body:  "If you received this notification it means you are configured correctly",
	})

	if sent != 1 || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error, token might be invalid or it's problem on our side",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "send test notification"})
}
