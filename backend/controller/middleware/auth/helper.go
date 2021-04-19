package auth

import (
	"encoding/base64"
	"strings"

	"github.com/gin-gonic/gin"
)

func extractToken(context *gin.Context) string {
	// Get JWT from Header Authorization
	bearToken := context.Request.Header.Get("Authorization")
	if bearToken == "" {
		v := context.Request.Header["authorization"]
		if len(v) != 0 {
			bearToken = v[0]
		}
	}
	strArr := strings.Split(bearToken, " ")
	if len(strArr) >= 2 {
		return strArr[1]
	}

	return ""
}

func extractBasicHeader(context *gin.Context) (string, string) {
	// Get JWT from Header Authorization
	basicToken := context.Request.Header.Get("Authorization")
	if basicToken == "" {
		v := context.Request.Header["authorization"]
		if len(v) != 0 {
			basicToken = v[0]
		}
	}

	strArr := strings.Split(basicToken, " ")
	if len(strArr) >= 2 {
		userPass, err := base64.StdEncoding.DecodeString(strArr[1])
		if err != nil {
			return "", ""
		}
		userPassArr := strings.Split(string(userPass), ":")
		if len(userPassArr) != 2 {
			return "", ""
		}
		return userPassArr[0], userPassArr[1]
	}
	return "", ""
}
