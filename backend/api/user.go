package api

import (
	"backendSenior/model"
	"backendSenior/repository"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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
