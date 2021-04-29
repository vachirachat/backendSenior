package auth

import (
	auth_service "backendSenior/domain/service/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProxyMiddleware provide function for creating various middleware for verifying JWT Token
type ProxyMiddleware struct {
	authService *auth_service.ProxyAuth
}

// NewProxyMiddleware create ProxyMiddleware
func NewProxyMiddleware(authSvc *auth_service.ProxyAuth) *ProxyMiddleware {
	return &ProxyMiddleware{
		authService: authSvc,
	}
}

// AuthRequired is used for route that require login.
// It will set userId, role in the `gin.Context`
func (mw *ProxyMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, clientSecret := extractBasicHeader(c)
		if clientID == "" {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "invalid token format"})
			return
		}

		ok, err := mw.authService.VerifyCredentials(clientID, clientSecret)
		if !ok {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "invalid token"})
			return
		}
		if err != nil {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "verification error: " + err.Error()})
			return
		}

		c.Set(UserIdField, clientID)
		c.Set(UserRoleField, "admin") // HACK[ROAD]: proxy is admin, so it can access /proxy/ API
		c.Next()

	}
}
