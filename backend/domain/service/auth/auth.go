package auth

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/utills"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthService struct {
	UserRepository repository.UserRepository
}

var RESOURCES = map[string]string{utills.ROLEUSER: utills.ROLEUSER, utills.ROLEADMIN: utills.ROLEADMIN}
var SCOPES = []string{"view", "add", "edit", "query"}

func (auth AuthService) AuthMiddleware(resouce string, scope string) gin.HandlerFunc {
	return func(context *gin.Context) {
		auth.canAccessResource(context, resouce, scope)
	}
}

func (auth AuthService) canAccessResource(context *gin.Context, resouce string, scope string) {
	_, role, ok := auth.getSession(context)
	log.Println(">>>>>>  ", role)
	// if isAdmin(role) || utills.ADMIN_MODE {
	if isAdmin(role) {
		context.Writer.WriteHeader(http.StatusOK)
	} else {
		if !ok {
			context.Abort()
			context.Writer.WriteHeader(http.StatusUnauthorized)
			context.Writer.Write([]byte("Unauthorized: not login state"))
		}
		if !hasPermission(context, role, "view") {
			context.Abort()
			context.Writer.WriteHeader(http.StatusUnauthorized)
			context.Writer.Write([]byte("Unauthorized: no permission"))
		}
		if !auth.roleScopesHandler(context, role, scope) {
			context.Abort()
			context.Writer.WriteHeader(http.StatusUnauthorized)
			context.Writer.Write([]byte("Unauthorized:Role no permission"))
		}
		context.Writer.WriteHeader(http.StatusOK)
	}
}

func (auth AuthService) getSession(context *gin.Context) (string, string, bool) {
	if auth.hasSession(context) {
		session, err := context.Request.Cookie("SESSION_ID")
		if err != nil {
			log.Println("error getSession", err.Error())
			context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
			return "", "", false
		}
		cookie, err := auth.UserRepository.GetUserIdByToken(session.Value)
		if err != nil {
			log.Println("error getSession", err.Error())
			context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
			return "", "", false
		}
		role, err := auth.UserRepository.GetUserRole(cookie.UserID.Hex())
		if err != nil {
			log.Println("error getSession", err.Error())
			context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
			return "", "", false
		}
		return cookie.UserID.Hex(), role, true
	}
	return "", "", false
}

func (auth AuthService) hasSession(context *gin.Context) bool {
	session, err := context.Request.Cookie("SESSION_ID")
	if err == nil {
		cookie, err := auth.UserRepository.GetUserIdByToken(session.Value)
		if err == nil {
			userTimeExp, _ := time.Parse(time.RFC3339, cookie.TimeExpired)
			if !isSessionExpire(userTimeExp) {
				return true
			} else {
				context.Abort()
				context.Writer.WriteHeader(http.StatusUnauthorized)
				context.Writer.Write([]byte("Unauthorized: your login state is expire"))
			}
		}
	}
	return false
}

func (auth AuthService) roleScopesHandler(context *gin.Context, role string, scope string) bool {
	p := model.Permission{Resource: role, Scopes: []string{}}
	isAdmin := isAdmin(role)
	isLoginState := auth.hasSession(context)

	for _, s := range SCOPES {
		if hasPermissionWithAdminFlag(role, s, isAdmin) && isLoginState {
			p.Scopes = append(p.Scopes, s)
		}
	}
	log.Println(p.Scopes)
	exist, _ := utills.In_array(scope, p.Scopes)
	return exist
}

func isSessionExpire(timeExp time.Time) bool {
	return !timeExp.Before(time.Now())
}

func isAdmin(role string) bool {
	_, ok := RESOURCES[role]
	if !ok {
		return false
	}
	if RESOURCES[role] != "admin" {
		return false
	}
	return true
}

func hasPermissionWithAdminFlag(role string, scope string, isAdmin bool) bool {
	if isAdmin || (scope == "view" && !isAdminResource(role)) {
		return true
	}
	return false
}

func isAdminResource(role string) bool {
	adminResource := []string{"admin"}
	for _, ar := range adminResource {
		if role == ar {
			return true
		}
	}
	return false
}

func hasPermission(c *gin.Context, role string, scope string) bool {
	isAdmin := isAdmin(role)
	if isAdmin ||
		(scope == "view" && !isAdminResource(role)) ||
		(role == "user" && scope == "add") {
		return true
	}
	return false
}

// func (auth AuthService) accessibleResourceHandler(context *gin.Context, role string, scope string) bool {
// 	var permissions = []model.Permission{}
// 	isAdmin := isAdmin(role)
// 	for _, r := range RESOURCES {
// 		p := model.Permission{Resource: r, Scopes: []string{}}
// 		isLoginState := auth.hasSession(context)
// 		for _, s := range SCOPES {
// 			// if utills.ADMIN_MODE || isLoginState && hasPermissionWithAdminFlag(context, r, s, isAdmin) {
// 			if isLoginState && hasPermissionWithAdminFlag(r, s, isAdmin) {
// 				p.Scopes = append(p.Scopes, s)
// 			}
// 		}
// 		permissions = append(permissions, p)
// 	}

// }
