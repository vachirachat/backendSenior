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

// AlternativeAuth is used so that proxy can access API as well as user, it DOES NOT enforce auth
// i.e. no 401 error when no token
func (mw *ProxyMiddleware) AlternativeAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, clientSecret := extractBasicHeader(c)
		if clientID == "" {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "invalid token format"})
			return
		}

		ok, err := mw.authService.VerifyCredentials(clientID, clientSecret)

		if ok && err == nil {
			c.Set(UserIdField, clientID)
			c.Set(UserRoleField, "proxy")
			c.Next()
		}

	}
}
