package middleware

import (
	"fmt"
	"net/http"
	"proxySenior/domain/service"

	"github.com/gin-gonic/gin"
)

const (
	UserIdField = "userId"
)

// AuthMiddleware provide function for creating various middleware for verifying JWT Token
type AuthMiddleware struct {
	authService *service.DelegateAuthService
}

// NewJWTMiddleware create JWTMiddleware
func NewAuthMiddleware(authService *service.DelegateAuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// AuthRequired is used for route that require login.
// It will set userId, role in the `gin.Context`
func (mw *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.Abort()
			fmt.Println("goo")
			c.JSON(http.StatusUnauthorized, gin.H{"status": "no token"})
			return
		}

		userID, err := mw.authService.Verify(token)
		if err != nil {
			c.Abort()
			fmt.Println("goo2")
			c.JSON(http.StatusUnauthorized, gin.H{"status": "invalid token: " + err.Error()})
			return
		}

		c.Set(UserIdField, userID)
		c.Next()

	}
}
