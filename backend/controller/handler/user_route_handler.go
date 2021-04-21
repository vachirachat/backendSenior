package route

import (
	authMw "backendSenior/controller/middleware/auth"
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"
	"backendSenior/utills"
	g "common/utils/ginutils"
	"fmt"
	"io/ioutil"

	"backendSenior/domain/model"
	"log"
	"net/http"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// UserRouteHandler is handler for route
type UserRouteHandler struct {
	userService    *service.UserService
	jwtService     *auth.JWTService
	authMiddleware *authMw.JWTMiddleware
	fileService    *service.FileService
}

func NewUserRouteHandler(
	userService *service.UserService,
	jwtService *auth.JWTService,
	authMiddleware *authMw.JWTMiddleware,
	fileService *service.FileService,
) *UserRouteHandler {
	return &UserRouteHandler{
		userService:    userService,
		jwtService:     jwtService,
		authMiddleware: authMiddleware,
		fileService:    fileService,
	}
}

// Mount make handle handle request for specified routerGroup
func (handler *UserRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("user", handler.userListHandler)
	routerGroup.PUT("me", handler.authMiddleware.AuthRequired(), handler.updateMyProfileHandler)
	routerGroup.DELETE("byid/:user_id", handler.deleteUserByIDHandler)
	routerGroup.POST("getuserbyemail", handler.getUserByEmail)

	//SignIN/UP API
	// routerGroup.GET("/token", handler.userTokenListHandler)
	routerGroup.POST("/login", handler.loginHandle)
	routerGroup.POST("/logout", handler.authMiddleware.AuthRequired(), handler.logoutHandle)
	// routerGroup.GET("/getalltoken", handler.getAllTokenHandle)

	routerGroup.POST("/login/:orgid/org", handler.loginOrgHandle)
	routerGroup.POST("/signup", handler.addUserSignUpHandeler)
	routerGroup.GET("/me", handler.authMiddleware.AuthRequired(), handler.getMeHandler)
	routerGroup.POST("/me/profile", handler.authMiddleware.AuthRequired(), g.InjectGin(handler.uploadProfileImage))

	// (for proxy)
	routerGroup.POST("/verify", handler.verifyToken)
	routerGroup.GET("/byid/:id", handler.getUserByIDHandler)

	routerGroup.GET("/byid/:id/profile", g.InjectGin(handler.getProfileImage))
}

func (handler *UserRouteHandler) getMeHandler(context *gin.Context) {
	id := context.GetString(authMw.UserIdField)

	user, err := handler.userService.GetUserByID(id)
	if err != nil {
		log.Println("error GetMe", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, user)
}

func (handler *UserRouteHandler) userListHandler(context *gin.Context) {
	var usersInfo model.UserInfo
	users, err := handler.userService.GetAllUsers()
	if err != nil {
		log.Println("error userListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	usersInfo.User = users
	context.JSON(http.StatusOK, usersInfo)
}

// for get user by id
func (handler *UserRouteHandler) getUserByIDHandler(context *gin.Context) {
	// Test
	log.Println("getUserByIDHandler by proxy")
	userID := context.Param("id")
	log.Println("userID")
	if !bson.IsObjectIdHex(userID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad user id"})
		return
	}
	user, err := handler.userService.GetUserByID(userID)
	if err != nil {
		log.Println("error GetUserByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusOK, user)
}

// GetUserByEmail for get user by id
func (handler *UserRouteHandler) getUserByEmail(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	user, err = handler.userService.GetUserByEmail(user.Email)
	if err != nil {
		log.Println("error GetUserByEmailHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, user)
}

func (handler *UserRouteHandler) updateMyProfileHandler(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error UpdateUserHandler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	myID := context.GetString(authMw.UserIdField)
	// Dont allow edit these field
	user.Email = ""
	user.Password = ""
	user.UserType = ""
	user.Room = nil
	user.Organize = nil
	// basically, currently, only name is editable

	err = handler.userService.UpdateUser(myID, user)
	if err != nil {
		log.Println("error UpdateUserHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *UserRouteHandler) deleteUserByIDHandler(context *gin.Context) {
	userID := context.Param("user_id")
	err := handler.userService.DeleteUserByID(userID)
	if err != nil {
		log.Println("error DeleteUserHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(http.StatusNoContent, gin.H{"status": "success"})
}

// // Edit user role
// func (handler *UserRouteHandler) editUseRoleHandler(context *gin.Context) {
// 	var credentials model.UserSecret
// 	err := context.ShouldBindJSON(&credentials)
// 	if err != nil {
// 		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
// 		return
// 	}
// 	err = handler.userService.EditUserRole(credentials)
// 	if err != nil {
// 		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
// 		return
// 	}
// 	context.JSON(http.StatusOK, gin.H{"status": "success"})
// }

func (handler *UserRouteHandler) loginHandle(context *gin.Context) {
	var credentials model.UserSecret
	err := context.ShouldBindJSON(&credentials)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}
	user, err := handler.userService.Login(credentials.Email, credentials.Password)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"status": err.Error()})
		return
	}

	tokenDetails, err := handler.jwtService.CreateToken(model.UserDetail{
		Role:   utills.ROLEUSER, // TODO: placeholder, implement role later
		UserId: user.UserID.Hex(),
	})

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success", "token": tokenDetails})
}

func (handler *UserRouteHandler) logoutHandle(context *gin.Context) {
	err := handler.jwtService.InvalidateToken(context.GetString(authMw.TokenField))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": "remove token error: " + err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

//func (handler *UserRouteHandler) getAllTokenHandle(context *gin.Context) {
//	tokens, err := handler.jwtService.GetAllToken()
//	if err != nil {
//		context.JSON(http.StatusBadRequest, gin.H{"status": "remove token error: " + err.Error()})
//		return
//	}
//
//	context.JSON(http.StatusOK, gin.H{"status": tokens})
//}

func (handler *UserRouteHandler) loginOrgHandle(context *gin.Context) {
	var credentials model.UserSecret
	err := context.ShouldBindJSON(&credentials)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}
	user, err := handler.userService.Login(credentials.Email, credentials.Password)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"status": err.Error()})
		return
	}

	orgID := context.Param("orgid")
	log.Println(orgID)
	err = handler.userService.IsUserInOrg(user, orgID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	tokenDetails, err := handler.jwtService.CreateToken(model.UserDetail{
		Role:   utills.ROLEUSER, // TODO: placeholder, implement role later
		UserId: user.UserID.Hex(),
	})

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success", "token": tokenDetails})
}

// Signup API
func (handler *UserRouteHandler) addUserSignUpHandeler(context *gin.Context) {
	var userPw model.UserWithPassword
	err := context.ShouldBindJSON(&userPw)
	if err != nil {
		log.Println("error AddUserSignUpHandeler user ShouldBindJSON", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	if userPw.Name == "" {
		context.JSON(http.StatusBadRequest, gin.H{"status": "not username specified"})
		return
	}
	user := *(*model.User)(unsafe.Pointer(&userPw))
	err = handler.userService.Signup(user)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"status": err.Error(),
		})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (handler *UserRouteHandler) verifyToken(context *gin.Context) {
	var body struct {
		Token string `json:"token"`
	}
	err := context.ShouldBindJSON(&body)
	if err != nil || body.Token == "" {
		context.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	claim, err := handler.jwtService.VerifyToken(body.Token)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": "verify error: " + err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"userId": claim.UserId,
	})
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
	// dunno why, but even image is PNG it still work
	c.Data(200, "image/jpeg", img)
	return nil
}

func (handler *UserRouteHandler) uploadProfileImage(c *gin.Context, req struct{}) error {
	userID := c.GetString(authMw.UserIdField)

	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("error getting form file: %w", err)
		return fmt.Errorf("error getting form file: %w", err)
	}

	f, err := file.Open()
	if err != nil {
		log.Printf("error opening file: %w", err)
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("error reading file: %w", err)
		return fmt.Errorf("error reading file: %w", err)
	}

	if err := handler.fileService.UploadProfileImage(userID, bytes); err != nil {
		log.Printf("error uploading image: %w", err)
		return fmt.Errorf("error uploading image: %w", err)
	}

	c.JSON(200, g.Response{
		Success: true,
		Message: "successfully uploaded profile",
		Data:    nil,
	})
	return nil
}
