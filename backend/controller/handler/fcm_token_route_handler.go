package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/dto"
	"backendSenior/domain/model"
	"backendSenior/domain/service"
	"backendSenior/utills"
	g "common/utils/ginutils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type FCMRouteHandler struct {
	notifService *service.NotificationService
	authMw       *auth.JWTMiddleware
	validate     *utills.StructValidator
}

func NewFCMRouteHandler(notif *service.NotificationService, authMw *auth.JWTMiddleware, validate *utills.StructValidator) *FCMRouteHandler {
	return &FCMRouteHandler{
		notifService: notif,
		authMw:       authMw,
		validate:     validate,
	}
}

func (handler *FCMRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.POST("/", handler.authMw.AuthRequired("user", "view"), g.InjectGin(handler.handleRegsiterDevice))
	routerGroup.GET("/", handler.authMw.AuthRequired("user", "view"), g.InjectGin(handler.handleGetUserDevices))
	routerGroup.DELETE("/", handler.authMw.AuthRequired("user", "view"), g.InjectGin(handler.handleUnregsiterDevice))
	routerGroup.POST("/ping", handler.authMw.AuthRequired("user", "view"), g.InjectGin(handler.handlePing))
	routerGroup.POST("/check-status", handler.authMw.AuthRequired("user", "view"), g.InjectGin(handler.checkTokenStatus))
	routerGroup.POST("/test-notification", handler.authMw.AuthRequired("user", "view"), g.InjectGin(handler.sendTestNotification))
}

func (handler *FCMRouteHandler) handleRegsiterDevice(c *gin.Context, input struct{ Body dto.FCMTokenDto }) error {
	userID := c.GetString(auth.UserIdField)

	// var body model.FCMToken
	// err := c.ShouldBindJSON(&body)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"status": err})
	// 	return
	// }
	b := input.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}
	tok, err := handler.notifService.GetTokenByID(b.Token)

	// found
	if tok.Token != "" {
		if tok.UserID.Hex() != userID {
			// c.JSON(http.StatusForbidden, gin.H{"status": "token already used by another user"})
			// return fmt.Errorf("token already used by another user")
			return g.NewError(403, "token already used by another user")
		}

		err = handler.notifService.RefreshDevice(tok.Token)
		if err != nil {
			// fmt.Println("[register device] refresh error", err)
			return g.NewError(500, "bad egister device refresh")
		}
		c.JSON(http.StatusOK, gin.H{"status": "refreshed"})
		return nil
	}

	// not found
	err = handler.notifService.RegisterDevice(userID, b.Token, b.DeviceName)
	if err != nil {
		// fmt.Println("[register device] error", err)
		// c.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return g.NewError(500, "register device error")
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

func (handler *FCMRouteHandler) handleGetUserDevices(c *gin.Context, req struct{}) error {
	userID := c.GetString(auth.UserIdField)

	tokens, err := handler.notifService.GetUserTokens(userID)

	if err != nil {
		if err.Error() == "not found" {
			c.JSON(http.StatusOK, []model.FCMToken{})
			return err
		} else {
			fmt.Println("[get user device] error", err)
			// c.JSON(http.StatusInternalServerError, gin.H{"status": err})
			return err
		}
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
	return nil
}

func (handler *FCMRouteHandler) handleUnregsiterDevice(c *gin.Context, input struct {
	Token string `json:"token" validate:"required,gt=0"`
}) error {
	userID := c.GetString(auth.UserIdField)

	b := input
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}
	// var body struct {
	// 	Token string `json:"token"`
	// }
	// err := c.ShouldBindJSON(&body)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"status": err})
	// 	return
	// }

	tokens, err := handler.notifService.GetUserTokens(userID)
	if err != nil {
		fmt.Println("[delete device] check permission error", err)
		// c.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return g.NewError(404, "bad check permission Token")
	}

	found := false
	for _, tok := range tokens {
		if tok.Token == b.Token {
			found = true
			break
		}
	}

	if !found {
		// c.JSON(http.StatusForbidden, gin.H{"status": "you don't own the token"})
		// return fmt.Errorf("token already used by another user")
		return g.NewError(403, "token already used by another user")
	}

	err = handler.notifService.DeleteDevice(b.Token)
	if err != nil {
		fmt.Println("[delete device] error", err)
		// c.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return err
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

// return status; <not found>, <owned>, <owned by other>
func (handler *FCMRouteHandler) checkTokenStatus(c *gin.Context, input struct {
	Token string `json:"token" validate:"required,gt=0"`
}) error {
	b := input
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	userID := c.GetString(auth.UserIdField)

	token, err := handler.notifService.GetTokenByID(b.Token)

	if err != nil {
		if err.Error() == "not found" {
			c.JSON(http.StatusOK, gin.H{"status": "not found"})
			return nil
		}
		fmt.Println("[check token status] error", err)
		// c.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return err
	}

	if token.UserID.Hex() == userID {
		c.JSON(http.StatusOK, gin.H{"status": "owned"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "owned by other"})
	}
	return nil
}

func (handler *FCMRouteHandler) handlePing(c *gin.Context, input struct {
	Token string `json:"token" validate:"required,gt=0"`
}) error {

	b := input
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	userID := c.GetString(auth.UserIdField)
	tok, err := handler.notifService.GetTokenByID(b.Token)
	if err != nil {
		fmt.Println("[notification ping] error", err)
		// c.JSON(http.StatusForbidden, gin.H{"status": "something went wrong"})
		return g.NewError(403, "something went wrong")
	} else if tok.UserID.Hex() != userID {
		// c.JSON(http.StatusForbidden, gin.H{"status": "not you token"})
		return g.NewError(403, "not your own token")
	}

	handler.notifService.SetLastSeenTime(b.Token, time.Now())
	c.JSON(http.StatusOK, gin.H{})
	return nil
}

func (handler *FCMRouteHandler) sendTestNotification(c *gin.Context, input struct {
	Token string `json:"token" validate:"required,gt=0"`
}) error {

	b := input
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	tok, err := handler.notifService.GetTokenByID(b.Token)
	if err != nil {
		return g.NewError(500, "something went wrong, ensure that token exists")
	}

	userID := c.GetString(auth.UserIdField)

	if tok.UserID.Hex() != userID {
		return g.NewError(403, "not your own token")
	}

	sent, err := handler.notifService.SendNotifications([]string{b.Token}, &model.Notification{
		Title: "Test notification",
		Body:  "If you received this notification it means you are configured correctly",
	})

	if sent != 1 || err != nil {
		return g.NewError(400, "error, token might be invalid or it's problem on our side")
	}

	c.JSON(http.StatusOK, gin.H{"status": "send test notification"})
	return nil
}
