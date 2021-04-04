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
	authMw          *auth.JWTMiddleware
}

// NewOrganizeRouteHandler create handler for Organize route
func NewOrganizeRouteHandler(organizeService *service.OrganizeService, authMw *auth.JWTMiddleware, userService *service.UserService, roomService *service.RoomService) *OrganizeRouteHandler {
	return &OrganizeRouteHandler{
		organizeService: organizeService,
		authMw:          authMw,
		userService:     userService,
		roomService:     roomService,
	}
}

//Mount make OrganizeRouteHandler handler request from specific `RouterGroup`
func (handler *OrganizeRouteHandler) Mount(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/:id", handler.getOrganizeByIDHandler)
	routerGroup.GET("/:id/org", handler.getOrganizeByNameHandler)
	routerGroup.POST("/", handler.authMw.AuthRequired(), handler.addOrganizeHandler)
	routerGroup.GET("/", handler.authMw.AuthRequired(), handler.getOrganizations)
	routerGroup.PUT("/:id", handler.editOrganizeNameHandler)
	routerGroup.DELETE("/:id", handler.deleteOrganizeByIDHandler)

	routerGroup.GET("/:id/member", handler.getOrganizationMembers)
	routerGroup.POST("/:id/member", handler.addMemberToOrganize)
	routerGroup.DELETE("/:id/member", handler.deleteMemberFromOrganize)

	routerGroup.GET("/:id/admin", handler.getOrganizationAdmins)
	routerGroup.POST("/:id/admin", handler.addAdminsToOrganize)
	routerGroup.DELETE("/:id/admin", handler.deleteAdminsFromOrganize)

	routerGroup.GET("/:id/room", handler.getOrgRooms)
}

func (handler *OrganizeRouteHandler) MountV2(rg *gin.RouterGroup) {

	rg.GET("/id/:id", handler.getOrganizeByIDHandler)
	rg.GET("/id/:id/org", handler.getOrganizeByNameHandler)
	rg.POST("/create-org", handler.authMw.AuthRequired(), handler.addOrganizeHandler)
	rg.GET("/", handler.authMw.AuthRequired(), handler.getOrganizations)
	rg.PUT("/id/:id", handler.editOrganizeNameHandler)
	rg.DELETE("/id/:id", handler.deleteOrganizeByIDHandler)

	rg.GET("/id/:id/member", handler.getOrganizationMembers)
	rg.POST("/id/:id/member", handler.addMemberToOrganize)
	rg.DELETE("/id/:id/member", handler.deleteMemberFromOrganize)

	rg.GET("/id/:id/admin", handler.getOrganizationAdmins)
	rg.POST("/id/:id/admin", handler.addAdminsToOrganize)
	rg.DELETE("/id/:id/admin", handler.deleteAdminsFromOrganize)

	rg.GET("/id/:id/room", handler.getOrgRooms)

	rg.POST("/find-org", g.InjectGin(handler.findOrgByName))
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

func (handler *OrganizeRouteHandler) getOrganizeByNameHandler(context *gin.Context) {
	OrganizeName := context.Param("id")
	Organize, err := handler.organizeService.GetOrganizeByName(OrganizeName)
	if err != nil {
		log.Println("error GetOrganizeByNameHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, Organize.OrganizeID)
}

// create an empty org, then the creator of the org is automatically invited to the org
func (handler *OrganizeRouteHandler) addOrganizeHandler(context *gin.Context) {
	var Organize model.Organize

	var orgID string
	isOK := false

	defer func() {
		if !isOK && orgID != "" {
			handler.organizeService.DeleteOrganizeByID(orgID)
		}
	}()

	err := context.ShouldBindJSON(&Organize)
	if err != nil {
		log.Println("error AddOrganizeHandeler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	orgID, err = handler.organizeService.AddOrganize(Organize)
	if err != nil {
		log.Println("error AddOrganizeHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	userID := context.GetString(auth.UserIdField)
	err = handler.organizeService.AddAdminToOrganize(orgID, []string{userID})
	if err != nil {
		log.Println("error AddOrganizeHandeler; invite self to room", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	isOK = true

	context.JSON(http.StatusCreated, gin.H{"status": "success", "orgId": orgID})
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

func (handler *OrganizeRouteHandler) getOrgRooms(c *gin.Context) {
	orgID := c.Param("id")
	if !bson.IsObjectIdHex(orgID) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad room ID"})
		return
	}

	roomIDs, err := handler.organizeService.GetOrgRoomIDs(orgID)
	if err != nil {
		fmt.Println("get org room error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	rooms, err := handler.roomService.GetRoomsByIDs(roomIDs)
	if err != nil {
		fmt.Println("get org room error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})

}

func (handler *OrganizeRouteHandler) findOrgByName(c *gin.Context, req struct {
	Body dto.FindOrgByNameDto
}) error {
	if orgs, err := handler.organizeService.FindOrgByName(req.Body); err != nil {
		return err
	} else {
		c.JSON(200, orgs[:20]) // limit don't show too much, privacy
		return nil
	}
}
