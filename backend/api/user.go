package api

import (
	"backendSenior/model"
	"backendSenior/repository"
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
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	usersInfo.User = users
	context.JSON(http.StatusOK, usersInfo)
}

// for get user by id
func (api UserAPI) GetUserByIDHandler(context *gin.Context) {
	userID := context.Param("user_id")
	user, err := api.UserRepository.GetUserByID(userID)
	if err != nil {
		log.Println("error GetUserByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusOK, user)
}

func (api UserAPI) AddUserHandeler(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error AddUserHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	err = api.UserRepository.AddUser(user)
	if err != nil {
		log.Println("error AddUserHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (api UserAPI) EditUserNameHandler(context *gin.Context) {
	var user model.User
	userID := context.Param("user_id")
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error EditProducNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	err = api.UserRepository.EditUserName(userID, user)
	if err != nil {
		log.Println("error EditProducNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (api UserAPI) DeleteUserByIDHandler(context *gin.Context) {
	userID := context.Param("user_id")
	err := api.UserRepository.DeleteUserByID(userID)
	if err != nil {
		log.Println("error DeleteUserHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
	context.JSON(http.StatusNoContent, gin.H{"message": "success"})
}

// Token

func (api UserAPI) UserTokenListHandler(context *gin.Context) {
	var usersTokenInfo model.UserTokenInfo
	usersTokens, err := api.UserRepository.GetAllUserToken()
	if err != nil {
		log.Println("error userListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusOK, token)
}

func (api UserAPI) LoginHandle(context *gin.Context) {
	// test cookie
	var userlogin model.UserLogin
	err := context.ShouldBindJSON(&userlogin)
	user, err := api.UserRepository.GetUserLogin(userlogin)
	if err != nil {
		log.Println("error LoginHandle", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}
	sessionCookie := &http.Cookie{Name: "SESSION_ID", Value: usertoken.Token, HttpOnly: false, Expires: time.Now().Add(30 * time.Minute), Path: "/"}
	http.SetCookie(context.Writer, sessionCookie)
	context.JSON(http.StatusOK, user)
}

//signUp
func (api UserAPI) AddUserSignUpHandeler(context *gin.Context) {
	var user model.User
	var userSecret model.UserLogin
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error AddUserLoginHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	userSecret.Password = user.Password
	userSecret.Username = user.Email
	log.Println(userSecret)
	err = api.UserRepository.AddUser(user)
	if err != nil {
		log.Println("error AddUserLoginHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	err = api.UserRepository.AddUserSecrect(userSecret)
	if err != nil {
		log.Println("error AddUserLoginHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

// func (api UserAPI) AddUserTokenHandeler(context *gin.Context, username string) {
// 	var usertoken model.UserToken
// 	err := api.UserRepository.AddToken(usertoken)
// 	if err != nil {
// 		log.Println("error AddUserTokenHandeler", err.Error())
// 		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
// 		return
// 	}
// 	context.JSON(http.StatusCreated, gin.H{"status": "success"})
// }
