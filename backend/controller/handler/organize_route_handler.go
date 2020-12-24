package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/model"
	"backendSenior/domain/service"
	"backendSenior/utills"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// OrganizeRouteHandler is Handler (controller) for Organize related route
type OrganizeRouteHandler struct {
	organizeService *service.OrganizeService
	userService     *service.UserService
	authMw          *auth.JWTMiddleware
}

// NewOrganizeRouteHandler create handler for Organize route
func NewOrganizeRouteHandler(organizeService *service.OrganizeService, authMw *auth.JWTMiddleware, userService *service.UserService) *OrganizeRouteHandler {
	return &OrganizeRouteHandler{
		organizeService: organizeService,
		authMw:          authMw,
		userService:     userService,
	}
}

//Mount make OrganizeRouteHandler handler request from specific `RouterGroup`
func (handler *OrganizeRouteHandler) Mount(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/:id" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.getOrganizeByIDHandler)
	routerGroup.POST("/" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.addOrganizeHandler)
	routerGroup.GET("/", handler.authMw.AuthRequired(), handler.getOrganizations)
	routerGroup.PUT("/:id" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.editOrganizeNameHandler)
	routerGroup.DELETE("/:id" /*handler.authService.AuthMiddleware("object", "view"),*/, handler.deleteOrganizeByIDHandler)

	routerGroup.GET("/:id/member", handler.getOrganizationMembers)
	routerGroup.POST("/:id/member", handler.addMemberToOrganize)
	routerGroup.DELETE("/:id/member", handler.deleteMemberFromOrganize)

	routerGroup.GET("/:id/admin", handler.getOrganizationAdmins)
	routerGroup.POST("/:id/admin", handler.addAdminsToOrganize)
	routerGroup.DELETE("/:id/admin", handler.deleteAdminsFromOrganize)

}

// return array of User that is admin of the organization
func (handler *OrganizeRouteHandler) getOrganizationAdmins(context *gin.Context) {
	orgID := context.Param("id")
	if !bson.IsObjectIdHex(orgID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad OrganizeID"})
		return
	}

	org, err := handler.organizeService.GetOrganizeById(orgID)
	if err != nil {
		fmt.Println("[getOrgAdmins]", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	adminUsers, err := handler.userService.GetUsersByIDs(utills.ToStringArr(org.Admins))
	if err != nil {
		fmt.Println("[getRoomProxies]", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"admins": adminUsers,
	})
}

func (handler *OrganizeRouteHandler) getOrganizationMembers(context *gin.Context) {
	orgID := context.Param("id")
	if !bson.IsObjectIdHex(orgID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad OrganizeID"})
		return
	}

	org, err := handler.organizeService.GetOrganizeById(orgID)
	if err != nil {
		log.Println("error GetOrganizeByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	users, err := handler.userService.GetUsersByIDs(utills.ToStringArr(org.Members))
	if err != nil {
		fmt.Println("[getRoomProxies]", err)
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"members": users,
	})
}

func isInObjArr(id bson.ObjectId, arr []bson.ObjectId) bool {
	for _, x := range arr {
		if id == x {
			return true
		}
	}
	return false
}

func (handler *OrganizeRouteHandler) getOrganizations(context *gin.Context) {
	var OrganizesInfo model.OrganizeInfo
	isMe := context.Query("me") != ""
	var orgs []model.Organize
	var err error

	if isMe {
		userID := context.GetString(auth.UserIdField)
		orgs, err = handler.organizeService.GetUserOrganizations(userID)
	} else {
		orgs, err = handler.organizeService.GetAllOrganizes()
	}

	if err != nil {
		log.Println("error OrganizeListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	OrganizesInfo.Orgs = orgs
	context.JSON(http.StatusOK, OrganizesInfo)
}

func (handler *OrganizeRouteHandler) getOrganizeByIDHandler(context *gin.Context) {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad OrganizeID"})
		return
	}

	Organize, err := handler.organizeService.GetOrganizeById(OrganizeID)
	if err != nil {
		log.Println("error GetOrganizeByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, Organize)
}

func (handler *OrganizeRouteHandler) addOrganizeHandler(context *gin.Context) {
	var Organize model.Organize
	err := context.ShouldBindJSON(&Organize)
	if err != nil {
		log.Println("error AddOrganizeHandeler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	OrganizeID, err := handler.organizeService.AddOrganize(Organize)
	if err != nil {
		log.Println("error AddOrganizeHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success", "orgId": OrganizeID})
}

func (handler *OrganizeRouteHandler) editOrganizeNameHandler(context *gin.Context) {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad OrganizeID"})
		return
	}

	var Organize model.Organize
	err := context.ShouldBindJSON(&Organize)
	Organize.OrganizeID = ""

	if err != nil {
		log.Println("error EditOrganizeNametHandler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	err = handler.organizeService.EditOrganizeName(OrganizeID, Organize)

	if err != nil {
		log.Println("error EditOrganizeNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *OrganizeRouteHandler) deleteOrganizeByIDHandler(context *gin.Context) {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad OrganizeID"})
		return
	}

	err := handler.organizeService.DeleteOrganizeByID(OrganizeID)
	if err != nil {
		log.Println("error DeleteOrganizeHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Match with Socket-structure

//// -- JoinAPI -> getSession(Topic+#ID) -> giveUserSession
func (handler *OrganizeRouteHandler) addMemberToOrganize(context *gin.Context) {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad OrganizeID"})
		return
	}

	// use bson.ObjectID to validate when bind
	var body struct {
		UserIDs []bson.ObjectId `json:"userIDs"`
	}

	err := context.ShouldBindJSON(&body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	err = handler.organizeService.AddMemberToOrganize(OrganizeID, utills.ToStringArr(body.UserIDs))
	if err != nil {
		log.Println("error AddMemberToOrganize", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *OrganizeRouteHandler) deleteMemberFromOrganize(context *gin.Context) {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad OrganizeID"})
		return
	}

	// use bson.ObjectID to validate when bind
	var body struct {
		UserIDs []bson.ObjectId `json:"userIDs"`
	}

	err := context.ShouldBindJSON(&body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	err = handler.organizeService.DeleteMemberFromOrganize(OrganizeID, utills.ToStringArr(body.UserIDs))

	if err != nil {
		log.Println("error DeleteOrganizeHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *OrganizeRouteHandler) addAdminsToOrganize(context *gin.Context) {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad OrganizeID"})
		return
	}

	// use bson.ObjectID to validate when bind
	var body struct {
		UserIDs []bson.ObjectId `json:"userIDs"`
	}

	err := context.ShouldBindJSON(&body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	err = handler.organizeService.AddAdminToOrganize(OrganizeID, utills.ToStringArr(body.UserIDs))
	if err != nil {
		log.Println("error AddMemberToOrganize", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *OrganizeRouteHandler) deleteAdminsFromOrganize(context *gin.Context) {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad OrganizeID"})
		return
	}

	// use bson.ObjectID to validate when bind
	var body struct {
		UserIDs []bson.ObjectId `json:"userIDs"`
	}

	err := context.ShouldBindJSON(&body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	err = handler.organizeService.DeleteAdminFromOrganize(OrganizeID, utills.ToStringArr(body.UserIDs))

	if err != nil {
		log.Println("error DeleteOrganizeHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}
