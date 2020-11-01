package auth

import (
	"backendSenior/repository"
	"fmt"
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
		cookie, ok := getSessionCookie(context)

		fmt.Println("authMiddleware")
		fmt.Println(cookie)
		fmt.Println(ok)

		if !ok {
			context.Abort()
			context.Writer.WriteHeader(http.StatusUnauthorized)
			context.Writer.Write([]byte("Unauthorized"))
			return
		}

	}
}

// func isInSession()
func getSessionCookie(c *gin.Context) (string, bool) {
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
