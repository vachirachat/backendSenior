package auth

import (
	"backendSenior/domain/service"
	auth_service "backendSenior/domain/service/auth"
	"backendSenior/utills"
	"github.com/ahmetb/go-linq/v3"
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

var _AccessLevel = []string{"user", "admin", "proxy"}

// AuthRequired is used for route that require login.
// It will set userId, role in the `gin.Context`
func (mw *JWTMiddleware) AuthRequired(requiredRole string, scope string) gin.HandlerFunc {
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

		minLevel := linq.From(_AccessLevel).IndexOf(func(i interface{}) bool {
			return i == requiredRole
		})
		curLevel := linq.From(_AccessLevel).IndexOf(func(i interface{}) bool {
			return i == claim.Role
		})

		if curLevel < minLevel {
			c.Abort()
			c.JSON(403, gin.H{"status": "not enough access level, expected at least" + requiredRole})
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
			return
		}

		roomID := c.Param(paramName)
		if !bson.IsObjectIdHex(roomID) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "bad param name"})
			c.Abort()
			return
		}

		userID := c.GetString(UserIdField)
		room, _ := mw.roomService.GetRoomByID(roomID)

		ok, _ := utills.In_array(bson.ObjectIdHex(userID), room.ListAdmin)
		if !ok {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "not in Room's Admin"})
			return
		}
		c.Next()
	}
}

func (mw *JWTMiddleware) IsInRoomMiddleWare(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(UserRoleField)
		if role == "proxy" {
			c.Next()
			return
		}

		roomID := c.Param(paramName)
		if !bson.IsObjectIdHex(roomID) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "bad param name"})
			c.Abort()
			return
		}

		userID := c.GetString(UserIdField)
		room, _ := mw.roomService.GetRoomByID(roomID)

		ok, _ := utills.In_array(bson.ObjectIdHex(userID), room.ListUser)
		if !ok {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "not in Room's User"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// IsInRoomMiddleWareQuery is same as IsInRoomMiddleware, but use query string to check
func (mw *JWTMiddleware) IsInRoomMiddleWareQuery(queryName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(UserRoleField)
		if role == "proxy" {
			c.Next()
			return
		}

		roomID := c.Query(queryName)
		if !bson.IsObjectIdHex(roomID) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "bad query string"})
			c.Abort()
			return
		}

		userID := c.GetString(UserIdField)
		room, _ := mw.roomService.GetRoomByID(roomID)

		ok, _ := utills.In_array(bson.ObjectIdHex(userID), room.ListUser)
		if !ok {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "not in Room's User"})
			return
		}
		c.Next()
	}
}

func (mw *JWTMiddleware) IsOrgAdmitMiddleWare(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(UserRoleField)
		if role == "proxy" {
			c.Next()
			return
		}

		orgID := c.Param(paramName)
		if !bson.IsObjectIdHex(orgID) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "bad param name"})
			c.Abort()
			return
		}

		userID := c.GetString(UserIdField)
		org, _ := mw.orgService.GetOrganizeById(orgID)

		ok, _ := utills.In_array(bson.ObjectIdHex(userID), org.Admins)
		if !ok {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "not in Org's Admin"})
			return
		}
		c.Next()
	}
}

func (mw *JWTMiddleware) IsInOrgMiddleWare(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(UserRoleField)
		if role == "proxy" {
			c.Next()
			return
		}

		orgID := c.Param(paramName)
		if !bson.IsObjectIdHex(orgID) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "bad param name"})
			c.Abort()
			return
		}

		userID := c.GetString(UserIdField)
		org, _ := mw.orgService.GetOrganizeById(orgID)

		ok, _ := utills.In_array(bson.ObjectIdHex(userID), org.Members)
		log.Println("middleware Org>>>")
		log.Println(userID, ok)
		if !ok {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{"status": "not in Org's User"})
			return
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
