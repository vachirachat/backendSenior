package route

import (
	authMw "backendSenior/controller/middleware/auth"
	"backendSenior/domain/dto"
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"
	"backendSenior/utills"
	g "common/utils/ginutils"
	"fmt"
	"io/ioutil"

	"backendSenior/domain/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// UserRouteHandler is handler for route
type UserRouteHandler struct {
	userService    *service.UserService
	jwtService     *auth.JWTService
	authMiddleware *authMw.JWTMiddleware
	fileService    *service.FileService
	validate       *utills.StructValidator
}

func NewUserRouteHandler(
	userService *service.UserService,
	jwtService *auth.JWTService,
	authMiddleware *authMw.JWTMiddleware,
	fileService *service.FileService,
	validate *utills.StructValidator,

) *UserRouteHandler {
	return &UserRouteHandler{
		userService:    userService,
		jwtService:     jwtService,
		authMiddleware: authMiddleware,
		fileService:    fileService,
		validate:       validate,
	}
}

// Mount make handle handle request for specified routerGroup
func (handler *UserRouteHandler) Mount(routerGroup *gin.RouterGroup) {

	routerGroup.GET("/me", handler.authMiddleware.AuthRequired("user", "view"), g.InjectGin(handler.getMeHandler))
	routerGroup.PUT("/me", handler.authMiddleware.AuthRequired("user", "edit"), g.InjectGin(handler.updateMyProfileHandler))
	routerGroup.POST("/me/profile", handler.authMiddleware.AuthRequired("user", "add"), g.InjectGin(handler.uploadProfileImage))

	routerGroup.GET("/byid/:id", handler.authMiddleware.AuthRequired("user", "view"), g.InjectGin(handler.getUserByIDHandler))
	routerGroup.DELETE("byid/:id", handler.authMiddleware.AuthRequired("user", "edit"), g.InjectGin(handler.deleteUserByIDHandler))
	routerGroup.GET("/byid/:id/profile", handler.authMiddleware.AuthRequired("user", "view"), g.InjectGin(handler.getProfileImage))

	// routerGroup.POST("/getuserbyemail", g.InjectGin(handler.getUserByEmail))

	//SignIN/UP API
	routerGroup.POST("/login", g.InjectGin(handler.loginHandle))
	routerGroup.POST("/login/:orgid/org", g.InjectGin(handler.loginOrgHandle))
	routerGroup.POST("/logout", handler.authMiddleware.AuthRequired("user", "edit"), g.InjectGin(handler.logoutHandle))
	routerGroup.POST("/signup", g.InjectGin(handler.addUserSignUpHandeler))

	// (for proxy)
	routerGroup.POST("/verify", g.InjectGin(handler.verifyToken))

	// Debug
	// routerGroup.GET("/getalltoken", handler.getAllTokenHandle)
	routerGroup.GET("/user", g.InjectGin(handler.userListHandler))
}

func (handler *UserRouteHandler) getMeHandler(context *gin.Context, req struct{}) error {
	id := context.GetString(authMw.UserIdField)

	user, err := handler.userService.GetUserByID(id)
	if err != nil {
		return g.NewError(404, "bad Get userid")
	}
	context.JSON(http.StatusOK, user)
	return nil
}

// Note > Edit: only email/name/
func (handler *UserRouteHandler) updateMyProfileHandler(context *gin.Context, input struct{ Body dto.UpdateMeDto }) error {
	myID := context.GetString(authMw.UserIdField)

	b := input.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	err = handler.userService.UpdateUser(myID, b.ToUser())
	if err != nil {
		return g.NewError(404, "bad update userid")
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

func (handler *UserRouteHandler) uploadProfileImage(c *gin.Context, req struct{}) error {
	userID := c.GetString(authMw.UserIdField)

	file, err := c.FormFile("image")
	if err != nil {
		return fmt.Errorf("error getting form file: %w", err)
	}

	f, err := file.Open()
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if err := handler.fileService.UploadProfileImage(userID, bytes); err != nil {
		return fmt.Errorf("error uploading image: %w", err)
	}

	c.JSON(200, g.Response{
		Success: true,
		Message: "successfully uploaded profile",
		Data:    nil,
	})
	return nil
}

// for get user by id
func (handler *UserRouteHandler) getUserByIDHandler(context *gin.Context, req struct{}) error {
	// Test
	userID := context.Param("id")
	if !bson.IsObjectIdHex(userID) {
		return g.NewError(400, "bad user id in path")
	}
	user, err := handler.userService.GetUserByID(userID)
	if err != nil {
		return g.NewError(404, "bad user id")
	}
	context.JSON(http.StatusOK, user)
	return nil
}

func (handler *UserRouteHandler) deleteUserByIDHandler(context *gin.Context, req struct{}) error {
	id := context.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad user id in path")
	}
	err := handler.userService.DeleteUserByID(id)
	if err != nil {
		return g.NewError(404, "bad delete user id in path")
	}
	context.JSON(http.StatusNoContent, gin.H{"status": "success"})
	return nil
}

func (handler *UserRouteHandler) getProfileImage(c *gin.Context, req struct{}) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad user id in param")
	}

	isThumbnail := c.Query("thumbnail") == "true"

	img, err := handler.fileService.GetProfileImage(id, isThumbnail)
	if err != nil {
		return err
	}

	c.Header("Content-Disposition", "inline")
	c.Header("Content-Length", fmt.Sprint(len(img)))
	c.Data(200, "image/jpeg", img)
	return nil
}

// GetUserByEmail for get user by id
// func (handler *UserRouteHandler) getUserByEmail(context *gin.Context, input struct {
// 	Body struct {
// 		Email string `json:"email" validate:"required,gt=0,email" `
// 	}
// }) error {
// 	// var user model.User
// 	// err := context.ShouldBindJSON(&user)
// 	b := input.Body
// 	err := handler.validate.ValidateStruct(b)
// 	if err != nil {
// 		return g.NewError(400, "bad Body Email ")
// 	}

// 	user, err := handler.userService.GetUserByEmail(b.Email)
// 	if err != nil {
// 		return g.NewError(404, "Not found user")
// 	}
// 	context.JSON(http.StatusOK, user)
// 	return nil
// }

func (handler *UserRouteHandler) loginHandle(context *gin.Context, input struct{ Body dto.CreateUserSecret }) error {
	// var credentials model.UserSecret
	// err := context.ShouldBindJSON(&credentials)
	// if err != nil {
	// 	context.JSON(http.StatusBadRequest, gin.H{"status": "error"})
	// 	return
	// }
	b := input.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}
	user, err := handler.userService.Login(b)
	if err != nil {
		return g.NewError(404, "Not found user")
	}

	tokenDetails, err := handler.jwtService.CreateToken(model.UserDetail{
		Role:   utills.ROLEUSER, // TODO: placeholder, implement role later
		UserId: user.UserID.Hex(),
	})

	if err != nil {
		return g.NewError(406, "Wrong format role/userid token")
	}
	context.JSON(http.StatusOK, gin.H{"status": "success", "token": tokenDetails})
	return nil
}

func (handler *UserRouteHandler) loginOrgHandle(context *gin.Context, input struct{ Body dto.CreateUserSecret }) error {
	orgID := context.Param("orgid")
	if !bson.IsObjectIdHex(orgID) {
		return g.NewError(400, "bad user id in param")
	}
	b := input.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	user, err := handler.userService.Login(b)
	if err != nil {
		return g.NewError(404, "Not found user")
	}

	err = handler.userService.IsUserInOrg(user, orgID)
	if err != nil {
		return g.NewError(404, "Not found Org user")
	}

	tokenDetails, err := handler.jwtService.CreateToken(model.UserDetail{
		Role:   utills.ROLEUSER, // TODO: placeholder, implement role later
		UserId: user.UserID.Hex(),
	})

	if err != nil {
		return g.NewError(406, "Wrong format role/userid token")
	}
	context.JSON(http.StatusOK, gin.H{"status": "success", "token": tokenDetails})
	return nil
}

func (handler *UserRouteHandler) logoutHandle(context *gin.Context, req struct{}) error {
	id := context.GetString(authMw.UserIdField)
	err := handler.jwtService.RemoveToken(id)
	if err != nil {
		return g.NewError(403, "Wrong RemoveToken role/userid")
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}

// Signup API
func (handler *UserRouteHandler) addUserSignUpHandeler(context *gin.Context, input struct{ Body dto.CreateUser }) error {
	b := input.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	err = handler.userService.Signup(b.ToUser())
	if err != nil {
		return g.NewError(403, "bad Signup")
	}

	context.JSON(http.StatusCreated, gin.H{"status": "success"})
	return nil
}

func (handler *UserRouteHandler) verifyToken(context *gin.Context, input struct {
	Body struct {
		Token string `json:"token" validate:"required,len=268,gt=0" `
	}
}) error {

	b := input.Body
	err := handler.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad body format")
	}

	claim, err := handler.jwtService.VerifyToken(b.Token)
	if err != nil {
		return g.NewError(406, "Wrong VerifyToken")
	}

	context.JSON(http.StatusOK, gin.H{
		"userId": claim.UserId,
	})
	return nil
}

func (handler *UserRouteHandler) userListHandler(context *gin.Context, req struct{}) error {
	var usersInfo model.UserInfo
	users, err := handler.userService.GetAllUsers()
	if err != nil {
		return g.NewError(404, "bad GetAllUsers")
	}
	usersInfo.User = users
	context.JSON(http.StatusOK, usersInfo)
	return nil
}

// func (handler *UserRouteHandler) getAllTokenHandle(context *gin.Context) {
// 	tokens, err := handler.jwtService.GetAllToken()
// 	if err != nil {
// 		return g.NewError(404, "bad GetAllToken")
// 	}

// 	context.JSON(http.StatusOK, gin.H{"status": tokens})
// }

// // Edit user role
// func (handler *UserRouteHandler) editUseRoleHandler(context *gin.Context) {
// 	var credentials model.UserSecret
// 	err := context.ShouldBindJSON(&credentials)
// 	if err != nil {
// 		return g.NewError(400, "bad credentials")
// 	}
// 	err = handler.userService.EditUserRole(credentials)
// 	if err != nil {
// 		return g.NewError(403, "bad EditUserRole")
// 	}
// 	context.JSON(http.StatusOK, gin.H{"status": "success"})
// }
