package auth

import (
	auth_service "backendSenior/domain/service/auth"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	UserIdField   = "userId"
	UserRoleField = "role"
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

type Permission struct {
	Resource string   `json:"resource" bson:"resource"`
	Scopes   []string `json:"scopes" bson:"scopes"`
}

var RESOURCES = []string{"admin", "user"}
var SCOPES = []string{"view", "add", "edit", "query"}

// AuthRequired is used for route that require login.
// It will set userId, role in the `gin.Context`
func (mw *JWTMiddleware) AuthRequired(resouce string, scope string) gin.HandlerFunc {
	// func (mw *JWTMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		// Fix Debug : Token
		// token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6IjcxNWIwOGNkLTY3M2MtNDU2Ni04ZGQyLWRmMDAwODlmOGRiMSIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MjIxODgzODM3Nywicm9sZSI6InVzZXIiLCJ1c2VyX2lkIjoiNjA3NWU5YWE0NzBhYWNjNGFkNDA3ZDkwIn0.3htMS-9PUO1ZyJaNc8KwZbGhv54prIYCOP5_vMFSo80"
		err := canAccessResource(c)
		if err != nil {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "access: " + err.Error()})
			return
		}
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
		c.Next()

	}
}

func isAdmin(resource string) bool {
	if resource == "admin" {
		return true
	}
	return false
}

func isAdminResource(resource string) bool {
	adminResource := []string{"admin"}
	for _, ar := range adminResource {
		if resource == ar {
			return true
		}
	}
	return false
}

func hasPermission(c *gin.Context, resource string, scope string) bool {
	if isAdmin(resource) ||
		(scope == "view" && !isAdminResource(resource)) ||
		(scope == "add" && !isAdminResource(resource)) ||
		(scope == "edit" && !isAdminResource(resource)) {
		return true
	}
	return false
}

func canAccessResource(c *gin.Context) error {
	resource := c.Param("resource")
	if resource == "admin" || resource == "user" {
		return nil
	} else {
		if !hasPermission(c, resource, "view") {
			return errors.New("Unauthorized: no permission")
		} else {
			return nil
		}
	}
	return nil
}
