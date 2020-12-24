package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func extractToken(context *gin.Context) string {
	// Get JWT from Header Authorization
	bearToken := context.Request.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
