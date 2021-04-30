package auth

import (
	"backendSenior/domain/service"
	auth_service "backendSenior/domain/service/auth"
	"backendSenior/utills"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

const (
	UserIdField   = "userId"
	UserRoleField = "role"
	TokenField    = "token"
)

// JWTMiddleware provide function for creating various middleware for verifying JWT Token
type JWTMiddleware struct {
	jwtService  *auth_service.JWTService
	roomService *service.RoomService
	orgService  *service.OrganizeService
}

// NewJWTMiddleware create JWTMiddleware
func NewJWTMiddleware(authSvc *auth_service.JWTService, roomSvc *service.RoomService, orgSvc *service.OrganizeService) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService:  authSvc,
		roomService: roomSvc,
		orgService:  orgSvc,
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
	// func (mw *JWTMiddleware) AlternativeAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// HACK[ROAD]: if other middleware already set UserId, Role field, then skip
		if c.GetString(UserIdField) != "" {
			return
		}

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

		log.Println("claim.Role", claim.Role)
		if !hasPermission(claim.Role, scope) {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "cannot access"})
			return
		}

		c.Set(UserIdField, claim.UserId)
		c.Set(UserRoleField, claim.Role)
		c.Set(TokenField, token)
		c.Next()

	}
}

func (mw *JWTMiddleware) IsRoomAdmitMiddleWare(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {

		role := c.GetString(UserRoleField)
		if role == "proxy" {
			c.Next()
		}

		roomID := c.Param(paramName)
		if !bson.IsObjectIdHex(roomID) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "bad param name"})
			c.Abort()
		}

		userID := c.GetString(UserIdField)
		room, _ := mw.roomService.GetRoomByID(roomID)

		ok, _ := utills.In_array(bson.ObjectIdHex(userID), room.ListAdmin)
		if !ok {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "not in Room's Admin"})
			c.Abort()
		}
		c.Next()
	}
}

func (mw *JWTMiddleware) IsInRoomMiddleWare(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(UserRoleField)
		if role == "proxy" {
			c.Next()
		}

		roomID := c.Param(paramName)
		if !bson.IsObjectIdHex(roomID) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "bad param name"})
			c.Abort()
		}

		userID := c.GetString(UserIdField)
		room, _ := mw.roomService.GetRoomByID(roomID)

		ok, _ := utills.In_array(bson.ObjectIdHex(userID), room.ListUser)
		if !ok {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "not in Room's User"})
			c.Abort()
		}
		c.Next()
	}
}

func (mw *JWTMiddleware) IsOrgAdmitMiddleWare(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(UserRoleField)
		if role == "proxy" {
			c.Next()
		}

		orgID := c.Param(paramName)
		if !bson.IsObjectIdHex(orgID) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "bad param name"})
			c.Abort()
		}

		userID := c.GetString(UserIdField)
		org, _ := mw.orgService.GetOrganizeById(orgID)

		ok, _ := utills.In_array(bson.ObjectIdHex(userID), org.Admins)
		if !ok {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "not in Org's Admin"})
			c.Abort()
		}
		c.Next()
	}
}

func (mw *JWTMiddleware) IsInOrgMiddleWare(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(UserRoleField)
		if role == "proxy" {
			c.Next()
		}

		orgID := c.Param(paramName)
		if !bson.IsObjectIdHex(orgID) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "bad param name"})
			c.Abort()
		}

		userID := c.GetString(UserIdField)
		org, _ := mw.orgService.GetOrganizeById(orgID)

		ok, _ := utills.In_array(bson.ObjectIdHex(userID), org.Members)
		log.Println("middleware Org>>>")
		log.Println(userID, ok)
		if !ok {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "not in Org's User"})
			c.Abort()
		}
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

func hasPermission(resource string, scope string) bool {
	if isAdmin(resource) ||
		(scope == "view" && !isAdminResource(resource)) ||
		(scope == "add" && !isAdminResource(resource)) ||
		(scope == "edit" && !isAdminResource(resource)) {
		return true
	}
	return false
}
