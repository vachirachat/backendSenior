package auth

import (
	auth_service "backendSenior/domain/service/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	UserIdField   = "userId"
	UserRoleField = "role"
	TokenField    = "token"
)

// JWTMiddleware provide function for creating various middleware for verifying JWT Token
type JWTMiddleware struct {
	jwtService *auth_service.JWTService
}

// NewJWTMiddleware create JWTMiddleware
func NewJWTMiddleware(authSvc *auth_service.JWTService) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService: authSvc,
	}
}

// AuthRequired is used for route that require login.
// It will set userId, role in the `gin.Context`
func (mw *JWTMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "no token"})
			return
		}

		claim, err := mw.jwtService.VerifyToken(token)
		if err != nil {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "invalid token: " + err.Error()})
			return
		}

		c.Set(UserIdField, claim.UserId)
		c.Set(UserRoleField, claim.Role)
		c.Set(TokenField, token)
		c.Next()

	}
}
