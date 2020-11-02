package auth

import (
	"backendSenior/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	UserRepository repository.UserRepository
}

type Permission struct {
	Resource string   `json:"resource" bson:"resource"`
	Scopes   []string `json:"scopes" bson:"scopes"`
}

var RESOURCES = []string{"r1", "r2", "r3"}
var SCOPES = []string{"view", "add", "edit", "query"}

func (auth Auth) AuthMiddleware(resouce string, scope string) gin.HandlerFunc {
	return func(context *gin.Context) {
		_, ok := getSession(context)
		if !ok {
			context.Abort()
			context.Writer.WriteHeader(http.StatusUnauthorized)
			context.Writer.Write([]byte("Unauthorized"))
			return
		}

	}
}

// func isInSession()
func getSession(c *gin.Context) (string, bool) {
	r := c.Request
	cookie, err := r.Cookie("SESSION_ID")
	if err == nil {
		userToken, err := repository.GetUserIdByToken(cookie.Value)
		if err == nil {
			return userToken.Email, true
		}
	}
	return "", false
}

// func CanAccessResource(c *gin.Context) {
// 	resource := c.Param("resource")
// 	if share.ADMIN_MODE || resource == "r1" || resource == "r2" {
// 		c.Writer.WriteHeader(http.StatusOK)
// 	} else {

// 		if _, ok := getSession(c); !ok {
// 			c.Abort()
// 			c.Writer.WriteHeader(http.StatusUnauthorized)
// 			c.Writer.Write([]byte("Unauthorized: not login state"))
// 		}
// 		if !hasPermission(c, resource, "view") {
// 			c.Abort()
// 			c.Writer.WriteHeader(http.StatusUnauthorized)
// 			c.Writer.Write([]byte("Unauthorized: no permission"))
// 		} else {
// 			c.Writer.WriteHeader(http.StatusOK)
// 		}
// 	}
// }

// func GetAccessibleResource(c *gin.Context) {
// 	var permissions = []Permission{}
// 	isAdmin := auth.IsAdmin(c)
// 	for _, r := range RESOURCES {
// 		p := Permission{Resource: r, Scopes: []string{}}
// 		_, isLoginState := db.GetSessionCookie(c)
// 		for _, s := range SCOPES {
// 			if share.ADMIN_MODE || r == "home" || r == "r1" || isLoginState && hasPermissionWithAdminFlag(c, r, s, isAdmin) {
// 				p.Scopes = append(p.Scopes, s)
// 			}
// 		}
// 		permissions = append(permissions, p)
// 	}
// 	c.Header("Content-Type", "application/json")
// 	c.Writer.WriteHeader(http.StatusOK)
// 	c.JSON(http.StatusOK, gin.H{
// 		"data": permissions,
// 	})
// }

// func GetResourceScopes(c *gin.Context) {
// 	resource := c.Param("resource")
// 	p := Permission{Resource: resource, Scopes: []string{}}
// 	isAdmin := isAdmin(c)
// 	for _, s := range SCOPES {
// 		if hasPermissionWithAdminFlag(c, resource, s, isAdmin) {
// 			p.Scopes = append(p.Scopes, s)
// 		}
// 	}
// 	c.Header("Content-Type", "application/json")
// 	c.Writer.WriteHeader(http.StatusOK)
// 	c.JSON(http.StatusOK, gin.H{
// 		"data": p,
// 	})
// }

// func hasPermission(c *gin.Context, resource string, scope string) bool {
// 	isAdmin := isAdmin(c)
// 	if isAdmin ||
// 		(scope == "view" && !isAdminResource(resource)) ||
// 		(resource == "user" && scope == "add") {
// 		return true
// 	}
// 	return false
// }

// func hasPermissionWithAdminFlag(c *gin.Context, resource string, scope string, isAdmin bool) bool {
// 	if isAdmin || (scope == "view" && !isAdminResource(resource)) {
// 		return true
// 	}
// 	return false
// }

// func isAdmin(c *gin.Context) bool {
// 	var userID interface{}
// 	var isAdmin = false
// 	userID, hasSession := db.GetSessionCookie(c)
// 	if hasSession {
// 		var u *user.UserSecretProfile
// 		u = user.SelectUserSecretProfileByID(userID.(string))
// 		isAdmin = u.IsAdmin
// 	}
// 	return isAdmin
// }

// func isAdminResource(resource string) bool {
// 	adminResource := []string{"r1", "r2"}
// 	for _, ar := range adminResource {
// 		if resource == ar {
// 			return true
// 		}
// 	}
// 	return false
// }
