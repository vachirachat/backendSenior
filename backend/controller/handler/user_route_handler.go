package route

import (
	"backendSenior/domain/service"

	"backendSenior/domain/model"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserRouteHandler is handler for route
type UserRouteHandler struct {
	userService *service.UserService
}

// NewUserRouteHandler create new route handler
func NewUserRouteHandler(userService *service.UserService) *UserRouteHandler {
	return &UserRouteHandler{
		userService: userService,
	}
}

// Mount make handle handle request for specified routerGroup
func (handler *UserRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("user", handler.userListHandler)
	routerGroup.PUT("user/updateuserprofile", handler.editUserNameHandler)
	routerGroup.DELETE("user/:user_id", handler.deleteUserByIDHandler)
	routerGroup.POST("getuserbyemail", handler.getUserByEmail)

	//SignIN/UP API
	routerGroup.GET("token", handler.userTokenListHandler)
	routerGroup.POST("login", handler.loginHandle)
	routerGroup.POST("signup", handler.addUserSignUpHandeler)
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
// func (handler *UserRouteHandler) getUserByIDHandler(context *gin.Context) {
// 	userID := context.Param("user_id")
// 	// user, err := handler.userService.GetUserByID(userID)
// 	if err != nil {
// 		log.Println("error GetUserByIDHandler", err.Error())
// 		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
// 		return
// 	}
// 	context.JSON(http.StatusOK, user)
// }

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

//for return roomidList of User
// func (handler *UserRouteHandler) getUserRoomByUserID(context *gin.Context) {
// 	var user model.User
// 	err := context.ShouldBindJSON(&user)
// 	userResult, err := handler.userService.GetUserByID(user.UserID)
// 	if err != nil {
// 		log.Println("error getUserRoomByUserID", err.Error())
// 		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
// 		return
// 	}
// 	roomIDList := userResult.Room
// 	log.Println(roomIDList)
// 	var roomNameList []string
// 	for _, s := range roomIDList {
// 		room, err := handler.userService.GetRoomWithRoomID(s)
// 		if err != nil {
// 			log.Println("error getUserRoomByUserID", err.Error())
// 			context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
// 			return
// 		}
// 		roomNameList = append(roomNameList, room.RoomName)
// 	}

// 	context.JSON(http.StatusOK, gin.H{"username": userResult.Name, "RoomIDList": userResult.Room, "RoomNameList": roomNameList})
// }

// AddUserHandeler api
func (handler *UserRouteHandler) addUserHandeler(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error AddUserHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	err = handler.userService.AddUser(user)
	if err != nil {
		log.Println("error AddUserHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

// EditUserNameHandler api
func (handler *UserRouteHandler) editUserNameHandler(context *gin.Context) {
	var user model.User
	// userID := context.Param("user_id")
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error EditUserNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	err = handler.userService.EditUserName(user.UserID.Hex(), user)
	if err != nil {
		log.Println("error EditUserNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *UserRouteHandler) updateUserHandler(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error UpdateUserHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	log.Println(user.UserID)
	err = handler.userService.UpdateUser(user.UserID.Hex(), user)
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

// Token
func (handler *UserRouteHandler) userTokenListHandler(context *gin.Context) {
	var usersTokenInfo model.UserTokenInfo
	usersTokens, err := handler.userService.UserTokenList()
	if err != nil {
		log.Println("error userListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	usersTokenInfo.UserToken = usersTokens
	context.JSON(http.StatusOK, usersTokenInfo)
}

func (handler *UserRouteHandler) getUserTokenByIDHandler(context *gin.Context) {
	userID := context.Param("token_id")
	token, err := handler.userService.GetUserTokenByID(userID)
	if err != nil {
		log.Println("error GetUserTokenByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, token)
}

type messageLogin struct {
	status string
	Email  string
	token  string
}

func (handler *UserRouteHandler) loginHandle(context *gin.Context) {
	// test cookie
	var credentials model.UserLogin
	err := context.ShouldBindJSON(&credentials)
	log.Println("Login Handle")
	log.Println(credentials)

	token, err := handler.userService.Login(credentials)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	m := messageLogin{"success", credentials.Email, token}
	context.JSON(http.StatusOK, m)
}

// Signup API
func (handler *UserRouteHandler) addUserSignUpHandeler(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error AddUserSignUpHandeler ShouldBindJSON", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}
	err = handler.userService.Signup(user)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"status": err.Error(),
		})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

//GetUserListSecrect

func (handler *UserRouteHandler) getUserListSecrect(context *gin.Context) {
	var usersInfo model.UserInfoSecrect
	users, err := handler.userService.GetAllUserSecret()
	if err != nil {
		log.Println("error userListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	usersInfo.UserLogin = users
	context.JSON(http.StatusOK, usersInfo)
}
