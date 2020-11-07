package api

import (
	"backendSenior/model"
	"backendSenior/repository"
	"backendSenior/utills"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
)

type UserAPI struct {
	UserRepository repository.UserRepository
}

func (api UserAPI) UserListHandler(context *gin.Context) {
	var usersInfo model.UserInfo
	users, err := api.UserRepository.GetAllUser()
	if err != nil {
		log.Println("error userListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	usersInfo.User = users
	context.JSON(http.StatusOK, usersInfo)
}

// for get user by id
// func (api UserAPI) GetUserByIDHandler(context *gin.Context) {
// 	userID := context.Param("user_id")
// 	// user, err := api.UserRepository.GetUserByID(userID)
// 	if err != nil {
// 		log.Println("error GetUserByIDHandler", err.Error())
// 		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
// 		return
// 	}
// 	context.JSON(http.StatusOK, user)
// }

// for get user by id
func (api UserAPI) GetUserByEmail(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	user, err = api.UserRepository.GetUserByEmail(user.Email)
	if err != nil {
		log.Println("error GetUserByEmailHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, user)
}

//for return roomidList of User
func (api UserAPI) GetUserRoomByUserID(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	userResult, err := api.UserRepository.GetUserByID(user.UserID)
	if err != nil {
		log.Println("error getUserRoomByUserID", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	roomIDList := userResult.Room
	log.Println(roomIDList)
	var roomNameList []string
	for _, s := range roomIDList {
		room, err := api.UserRepository.GetRoomWithRoomID(s)
		if err != nil {
			log.Println("error getUserRoomByUserID", err.Error())
			context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
			return
		}
		roomNameList = append(roomNameList, room.RoomName)
	}

	context.JSON(http.StatusOK, gin.H{"username": userResult.Name, "RoomIDList": userResult.Room, "RoomNameList": roomNameList})
}

func (api UserAPI) AddUserHandeler(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error AddUserHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	err = api.UserRepository.AddUser(user)
	if err != nil {
		log.Println("error AddUserHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (api UserAPI) EditUserNameHandler(context *gin.Context) {
	var user model.User
	// userID := context.Param("user_id")
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error EditUserNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	err = api.UserRepository.EditUserName(user.UserID, user)
	if err != nil {
		log.Println("error EditUserNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (api UserAPI) UpdateUserHandler(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error UpdateUserHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	log.Println(user.UserID)
	err = api.UserRepository.EditUserName(user.UserID, user)
	if err != nil {
		log.Println("error UpdateUserHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (api UserAPI) DeleteUserByIDHandler(context *gin.Context) {
	userID := context.Param("user_id")
	err := api.UserRepository.DeleteUserByID(userID)
	if err != nil {
		log.Println("error DeleteUserHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(http.StatusNoContent, gin.H{"status": "success"})
}

// Token
func (api UserAPI) UserTokenListHandler(context *gin.Context) {
	var usersTokenInfo model.UserTokenInfo
	usersTokens, err := api.UserRepository.GetAllUserToken()
	if err != nil {
		log.Println("error userListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	usersTokenInfo.UserToken = usersTokens
	context.JSON(http.StatusOK, usersTokenInfo)
}

func (api UserAPI) GetUserTokenByIDHandler(context *gin.Context) {
	userID := context.Param("token_id")
	token, err := api.UserRepository.GetUserTokenById(userID)
	if err != nil {
		log.Println("error GetUserTokenByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, token)
}

type messageLogin struct {
	Email string
	token string
}

func (api UserAPI) LoginHandle(context *gin.Context) {
	// test cookie
	var userlogin model.UserLogin
	err := context.ShouldBindJSON(&userlogin)
	log.Println("Login Handle")
	log.Println(userlogin)
	user, err := api.UserRepository.GetUserLogin(userlogin)
	if err != nil {
		log.Println("error LoginHandle", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	//Fix Check token
	var usertoken model.UserToken
	usertoken, err = api.UserRepository.GetUserTokenById(user.Username)

	// mean first_login or cookie is expired
	if err != nil {
		// if isexpied ?? implement
		log.Println("Pass IN if news token")
		usertoken.Email = user.Username
		usertoken.Token = ksuid.New().String()
		err = api.UserRepository.AddToken(usertoken)
		if err != nil {
			log.Println("error AddUserTokenHandeler", err.Error())
			context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
			return
		}
	}
	sessionCookie := &http.Cookie{Name: "SESSION_ID", Value: usertoken.Token, HttpOnly: false, Expires: time.Now().Add(30 * time.Minute), Path: "/"}
	http.SetCookie(context.Writer, sessionCookie)
	// map struct to return value
	m := messageLogin{user.Username, usertoken.Token}
	context.JSON(http.StatusOK, m)
}

// Signup API
func (api UserAPI) AddUserSignUpHandeler(context *gin.Context) {
	var user model.User
	var userSecret model.UserLogin
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error AddUserSignUpHandeler ShouldBindJSON", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	// isDuplicateEmail
	_, err = api.UserRepository.GetUserByEmail(user.Email)
	if err == nil {
		log.Println("error AddUserSignUpHandeler GetUserByEmail", err.Error())
		context.JSON(http.StatusOK, gin.H{"status": "already have this email"})
		return
	}

	// Add User to DB
	user.Password = utills.HashPassword(user.Password)
	err = api.UserRepository.AddUser(user)
	if err != nil {
		log.Println("error AddUserLoginHandeler AddUserToDB", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	// Add UserSecret to DB
	userSecret.Password = user.Password

	userSecret.Username = user.Email
	log.Println(userSecret)
	err = api.UserRepository.AddUserSecrect(userSecret)
	if err != nil {
		log.Println("error FrouthAddUserLoginHandeler AddToUserSecret", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

//GetUserListSecrect

func (api UserAPI) GetUserListSecrect(context *gin.Context) {
	var usersInfo model.UserInfoSecrect
	users, err := api.UserRepository.GetAllUserSecret()
	if err != nil {
		log.Println("error userListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	usersInfo.UserLogin = users
	context.JSON(http.StatusOK, usersInfo)
}
