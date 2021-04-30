package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/dto"
	"backendSenior/domain/model"
	"backendSenior/domain/service"
	"backendSenior/utills"
	g "common/utils/ginutils"
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
	roomService     *service.RoomService
	proxyService    *service.ProxyService
	authMw          *auth.JWTMiddleware
	validate        *utills.StructValidator
}

// NewOrganizeRouteHandler create handler for Organize route
func NewOrganizeRouteHandler(organizeService *service.OrganizeService, authMw *auth.JWTMiddleware, userService *service.UserService, proxyService *service.ProxyService, roomService *service.RoomService, validate *utills.StructValidator) *OrganizeRouteHandler {
	return &OrganizeRouteHandler{
		organizeService: organizeService,
		authMw:          authMw,
		userService:     userService,
		roomService:     roomService,
		proxyService:    proxyService,
		validate:        validate,
	}
}

//Mount make OrganizeRouteHandler handler request from specific `RouterGroup`
func (handler *OrganizeRouteHandler) Mount(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/:id", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInOrgMiddleWare("id"), g.InjectGin(handler.getOrganizeByIDHandler))
	routerGroup.PUT("/:id", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsOrgAdmitMiddleWare("id"), g.InjectGin(handler.editOrganizeNameHandler))
	routerGroup.DELETE("/:id", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsOrgAdmitMiddleWare("id"), g.InjectGin(handler.deleteOrganizeByIDHandler))

	routerGroup.POST("/", handler.authMw.AuthRequired("user", "add"), g.InjectGin(handler.addOrganizeHandler))
	routerGroup.GET("/", g.InjectGin(handler.getOrganizations))
	routerGroup.GET("/:id/org", g.InjectGin(handler.getOrganizeByNameHandler))

	routerGroup.GET("/:id/member", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInOrgMiddleWare("id"), g.InjectGin(handler.getOrganizationMembers))
	routerGroup.POST("/:id/member", handler.authMw.AuthRequired("user", "add"), handler.authMw.IsOrgAdmitMiddleWare("id"), g.InjectGin(handler.addMembersToOrganize))
	routerGroup.DELETE("/:id/member", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsOrgAdmitMiddleWare("id"), g.InjectGin(handler.deleteMemberFromOrganize))

	routerGroup.GET("/:id/admin", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInOrgMiddleWare("id"), g.InjectGin(handler.getOrganizationAdmins))
	routerGroup.POST("/:id/admin", handler.authMw.AuthRequired("user", "add"), handler.authMw.IsOrgAdmitMiddleWare("id"), g.InjectGin(handler.addAdminsToOrganize))
	routerGroup.DELETE("/:id/admin", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsOrgAdmitMiddleWare("id"), g.InjectGin(handler.deleteAdminsFromOrganize))

	routerGroup.GET("/:id/room", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInOrgMiddleWare("id"), g.InjectGin(handler.getOrgRooms))
}

func (handler *OrganizeRouteHandler) MountV2(rg *gin.RouterGroup) {

	rg.GET("/id/:id", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInOrgMiddleWare("id"), g.InjectGin(handler.getOrganizeByIDHandler))
	rg.GET("/id/:id/org", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInOrgMiddleWare("id"), g.InjectGin(handler.getOrganizeByNameHandler))
	rg.POST("/create-org", handler.authMw.AuthRequired("user", "add"), g.InjectGin(handler.addOrganizeHandler))
	rg.GET("/", handler.authMw.AuthRequired("user", "view"), g.InjectGin(handler.getOrganizations))
	rg.PUT("/id/:id", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsOrgAdmitMiddleWare("id"), g.InjectGin(handler.editOrganizeNameHandler))
	rg.DELETE("/id/:id", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsOrgAdmitMiddleWare("id"), g.InjectGin(handler.deleteOrganizeByIDHandler))

	rg.GET("/id/:id/member", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInOrgMiddleWare("id"), g.InjectGin(handler.getOrganizationMembers))
	rg.POST("/id/:id/member", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInOrgMiddleWare("id"), g.InjectGin(handler.addMembersToOrganize))
	rg.DELETE("/id/:id/member", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInOrgMiddleWare("id"), g.InjectGin(handler.deleteMemberFromOrganize))

	rg.GET("/id/:id/admin", handler.authMw.AuthRequired("user", "view"), handler.authMw.IsInOrgMiddleWare("id"), g.InjectGin(handler.getOrganizationAdmins))
	rg.POST("/id/:id/admin", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsOrgAdmitMiddleWare("id"), g.InjectGin(handler.addAdminsToOrganize))
	rg.DELETE("/id/:id/admin", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsOrgAdmitMiddleWare("id"), g.InjectGin(handler.deleteAdminsFromOrganize))

	rg.GET("/id/:id/room", handler.authMw.AuthRequired("user", "edit"), handler.authMw.IsOrgAdmitMiddleWare("id"), g.InjectGin(handler.getOrgRooms))

	rg.POST("/find-org", g.InjectGin(handler.findOrgByName))
}

func (handler *OrganizeRouteHandler) getOrganizeByIDHandler(context *gin.Context, req struct{}) error {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		return g.NewError(400, "Org id in path")
	}

	Organize, err := handler.organizeService.GetOrganizeById(OrganizeID)
	if err != nil {
		return g.NewError(404, "Org not found")
	}
	context.JSON(http.StatusOK, Organize)
	return nil
}

func (handler *OrganizeRouteHandler) editOrganizeNameHandler(context *gin.Context, req struct{ Body dto.OrgDto }) error {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		return g.NewError(400, "Org id in path")
	}
	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "Bad body format")
	}

	// var Organize model.Organize
	// err := context.ShouldBindJSON(&Organize)
	// Organize.OrganizeID = ""

	// if err != nil {
	// 	log.Println("error EditOrganizeNametHandler", err.Error())
	// 	context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
	// 	return
	// }
	err = handler.organizeService.EditOrganizeName(OrganizeID, b.ToOrg())
	if err != nil {
		return g.NewError(403, "fail to update org name")
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

func (handler *OrganizeRouteHandler) deleteOrganizeByIDHandler(context *gin.Context, req struct{}) error {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		return g.NewError(400, "Org id in path")
	}
	// Fix : Remove Proxies All related
	org, err := handler.organizeService.GetOrganizeById(OrganizeID)
	if err != nil {
		return g.NewError(403, "fail to delete org")
	}

	err = handler.proxyService.RemoveProxiseFromOrg(OrganizeID, utills.ToStringArr(org.Proxies))
	if err != nil {
		return g.NewError(403, "fail to delete org")
	}
	// Fix : Remove Proxies All related
	err = handler.organizeService.DeleteOrganizeByID(OrganizeID)
	if err != nil {
		// log.Println("error DeleteOrganizeHandler", err.Error())
		// context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return g.NewError(403, "fail to delete org")
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

// create an empty org, then the creator of the org is automatically invited to the org
func (handler *OrganizeRouteHandler) addOrganizeHandler(context *gin.Context, req struct{ Body dto.OrgDto }) error {
	var orgID string
	isOK := false

	defer func() {
		if !isOK && orgID != "" {
			handler.organizeService.DeleteOrganizeByID(orgID)
		}
	}()

	// err := context.ShouldBindJSON(&Organize)
	// if err != nil {
	// 	log.Println("error AddOrganizeHandeler", err.Error())
	// 	context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
	// 	return  err
	// }
	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "Bad body format")
	}

	orgID, err = handler.organizeService.AddOrganize(b.ToOrg())
	if err != nil {
		log.Println("error AddOrganizeHandeler", err.Error())
		// context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return err
	}

	userID := context.GetString(auth.UserIdField)
	err = handler.organizeService.AddAdminToOrganize(orgID, []string{userID})
	if err != nil {
		log.Println("error AddOrganizeHandeler; invite self to room", err.Error())
		// context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return err
	}

	isOK = true

	context.JSON(http.StatusCreated, gin.H{"status": "success", "orgId": orgID})
	return nil
}

func (handler *OrganizeRouteHandler) getOrganizations(context *gin.Context, req struct{}) error {
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
		// context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return err
	}
	OrganizesInfo.Orgs = orgs
	context.JSON(http.StatusOK, OrganizesInfo)
	return nil
}

// ID as name cause of :gin-gonic reg-ex limitation
func (handler *OrganizeRouteHandler) getOrganizeByNameHandler(context *gin.Context, req struct{}) error {
	OrganizeName := context.Param("id")
	Organize, err := handler.organizeService.GetOrganizeByName(OrganizeName)
	if err != nil {
		log.Println("error GetOrganizeByNameHandler", err.Error())
		// context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return err
	}
	context.JSON(http.StatusOK, Organize.OrganizeID)
	return nil
}

func (handler *OrganizeRouteHandler) getOrganizationMembers(context *gin.Context, req struct{}) error {
	orgID := context.Param("id")
	if !bson.IsObjectIdHex(orgID) {
		return g.NewError(400, "Org id in path")
	}

	org, err := handler.organizeService.GetOrganizeById(orgID)
	if err != nil {
		log.Println("error GetOrganizeByIDHandler", err.Error())
		// context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return err
	}

	users, err := handler.userService.GetUsersByIDs(utills.ToStringArr(org.Members))
	if err != nil {
		fmt.Println("[getRoomProxies]", err)
		// context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return err
	}

	context.JSON(http.StatusOK, gin.H{
		"members": users,
	})
	return nil
}

// Match with Socket-structure

//// -- JoinAPI -> getSession(Topic+#ID) -> giveUserSession
func (handler *OrganizeRouteHandler) addMembersToOrganize(context *gin.Context, req struct {
	Body struct {
		UserIDs []bson.ObjectId `json:"userIDs" validate:"required"`
	}
}) error {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		return g.NewError(400, "Org id in path")
	}

	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	err = handler.organizeService.AddMemberToOrganize(OrganizeID, utills.ToStringArr(b.UserIDs))
	if err != nil {
		log.Println("error AddMemberToOrganize", err.Error())
		// context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

func (handler *OrganizeRouteHandler) deleteMemberFromOrganize(context *gin.Context, req struct {
	Body struct {
		UserIDs []bson.ObjectId `json:"userIDs" validate:"required"`
	}
}) error {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		return g.NewError(400, "Org id in path")
	}

	// use bson.ObjectID to validate when bind
	// var body struct {
	// 	UserIDs []bson.ObjectId `json:"userIDs"`
	// }

	// err := context.ShouldBindJSON(&body)
	// if err != nil {
	// 	context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
	// 	return  err
	// }
	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}
	err = handler.organizeService.DeleteMemberFromOrganize(OrganizeID, utills.ToStringArr(req.Body.UserIDs))

	if err != nil {
		log.Println("error DeleteOrganizeHandler", err.Error())
		// context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

// return array of User that is admin of the organization
func (handler *OrganizeRouteHandler) getOrganizationAdmins(context *gin.Context, req struct{}) error {
	orgID := context.Param("id")
	if !bson.IsObjectIdHex(orgID) {
		return g.NewError(400, "Org id in path")
	}

	org, err := handler.organizeService.GetOrganizeById(orgID)
	if err != nil {
		fmt.Println("[getOrgAdmins]", err)
		// context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return err
	}

	adminUsers, err := handler.userService.GetUsersByIDs(utills.ToStringArr(org.Admins))
	if err != nil {
		fmt.Println("[getRoomProxies]", err)
		// context.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return err
	}

	context.JSON(http.StatusOK, gin.H{
		"admins": adminUsers,
	})
	return nil
}

func (handler *OrganizeRouteHandler) addAdminsToOrganize(context *gin.Context, req struct {
	Body struct {
		UserIDs []bson.ObjectId `json:"userIDs" validate:"required"`
	}
}) error {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		return g.NewError(400, "Org id in path")
	}

	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}
	err = handler.organizeService.AddAdminToOrganize(OrganizeID, utills.ToStringArr(req.Body.UserIDs))
	if err != nil {
		log.Println("error AddMemberToOrganize", err.Error())
		// context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

func (handler *OrganizeRouteHandler) deleteAdminsFromOrganize(context *gin.Context, req struct {
	Body struct {
		UserIDs []bson.ObjectId `json:"userIDs" validate:"required"`
	}
}) error {
	OrganizeID := context.Param("id")
	if !bson.IsObjectIdHex(OrganizeID) {
		return g.NewError(400, "Org id in path")
	}

	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}
	err = handler.organizeService.DeleteAdminFromOrganize(OrganizeID, utills.ToStringArr(req.Body.UserIDs))

	if err != nil {
		log.Println("error DeleteOrganizeHandler", err.Error())
		// context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return err
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

func (handler *OrganizeRouteHandler) getOrgRooms(c *gin.Context, req struct{}) error {
	orgID := c.Param("id")
	if !bson.IsObjectIdHex(orgID) {
		return g.NewError(400, "Org id in path")
	}

	roomIDs, err := handler.organizeService.GetOrgRoomIDs(orgID)
	if err != nil {
		fmt.Println("get org room error:", err)
		// c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return err
	}

	rooms, err := handler.roomService.GetRoomsByIDs(roomIDs)
	if err != nil {
		fmt.Println("get org room error:", err)
		// c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
	return nil

}

func (handler *OrganizeRouteHandler) findOrgByName(context *gin.Context, req struct {
	Body dto.FindOrgByNameDto
}) error {
	b := req.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	if orgs, err := handler.organizeService.FindOrgByName(b); err != nil {
		return err
	} else {
		if len(orgs) > 20 {
			orgs = orgs[:20] // limit don't show too much, privacy
		}
		context.JSON(200, orgs)
		return nil
	}
}
