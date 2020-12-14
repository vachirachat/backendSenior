package service

import (
	"backendSenior/domain/interface/repository"
	"errors"

	"backendSenior/domain/model"
	"backendSenior/utills"
	"log"

	"github.com/segmentio/ksuid"
)

// UserService provide access to user related functions
type UserService struct {
	userRepository repository.UserRepository
}

// NewUserService return instance of user service
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepo,
	}
}

//GetAllUsers return all users
func (service *UserService) GetAllUsers() ([]model.User, error) {
	users, err := service.userRepository.GetAllUser()
	return users, err
}

func (service *UserService) GetUserByID(userID string) (model.User, error) {
	user, err := service.userRepository.GetUserByID(userID)
	return user, err
}

// GetUserByEmail return user with specified email
func (service *UserService) GetUserByEmail(email string) (model.User, error) {
	user, err := service.userRepository.GetUserByEmail(email)
	return user, err
}

//for return roomidList of User
// func (api UserAPI) GetUserRoomByUserID(context *gin.Context) {
// 	var user model.User
// 	err := context.ShouldBindJSON(&user)
// 	userResult, err := service.userRepository.GetUserByID(user.UserID)
// 	if err != nil {
// 		log.Println("error getUserRoomByUserID", err.Error())
// 		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
// 		return
// 	}
// 	roomIDList := userResult.Room
// 	log.Println(roomIDList)
// 	var roomNameList []string
// 	for _, s := range roomIDList {
// 		room, err := service.userRepository.GetRoomWithRoomID(s)
// 		if err != nil {
// 			log.Println("error getUserRoomByUserID", err.Error())
// 			context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
// 			return
// 		}
// 		roomNameList = append(roomNameList, room.RoomName)
// 	}

// 	context.JSON(http.StatusOK, gin.H{"username": userResult.Name, "RoomIDList": userResult.Room, "RoomNameList": roomNameList})
// }

// AddUser create a user
func (service *UserService) AddUser(user model.User) error {

	err := service.userRepository.AddUser(user)
	return err
}

// EditUserName this actually update the whole user object
func (service *UserService) EditUserName(userID string, user model.User) error {
	err := service.userRepository.EditUserName(userID, user)
	return err
}

// UpdateUser update whole user
func (service *UserService) UpdateUser(userID string, user model.User) error {
	err := service.userRepository.EditUserName(userID, user)
	return err
}

// DeleteUserByID delete a user with specified ID
func (service *UserService) DeleteUserByID(userID string) error {
	err := service.userRepository.DeleteUserByID(userID)
	return err
}

// UserTokenList return all tokens from all users
func (service *UserService) UserTokenList() ([]model.UserToken, error) {
	userTokens, err := service.userRepository.GetAllUserToken()
	return userTokens, err
}

// GetUserTokenByID return all tokens of speicifed user
func (service *UserService) GetUserTokenByID(userID string) (model.UserToken, error) {
	token, err := service.userRepository.GetUserTokenById(userID)
	return token, err
}

type messageLogin struct {
	status string
	Email  string
	token  string
}

//Login find user with matching username, password, isAdmin, return token
// TODO WTF return
func (service *UserService) Login(credentials model.UserLogin) (string, error) {
	user, err := service.userRepository.GetUserLogin(credentials)

	//Fix Check token
	var usertoken model.UserToken
	usertoken, err = service.userRepository.GetUserTokenById(user.Email)
	log.Println(usertoken)

	// mean first_login or cookie is expired
	// if err != nil {
	// if isexpied ?? implement

	// generate new token
	log.Println("Pass IN if news token")
	usertoken.Email = user.Email
	usertoken.Token = ksuid.New().String()
	err = service.userRepository.AddToken(usertoken)

	// if generate error, error
	if err != nil {
		log.Println("error AddUserTokenHandeler", err.Error())
		return "", err
		// }
	}

	return usertoken.Token, nil
}

// Signup API
func (service *UserService) Signup(user model.User) error {
	_, err := service.userRepository.GetUserByEmail(user.Email)
	if err == nil {
		return errors.New("User already exists")
	}

	// Add User to DB
	user.Password = utills.HashPassword(user.Password)
	err = service.userRepository.AddUser(user)
	if err != nil {
		return err
	}

	err = service.userRepository.AddUserSecrect(model.UserLogin{
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return err
	}

	return nil
}

//GetAllUserSecret return secret of all user
func (service *UserService) GetAllUserSecret() ([]model.UserLogin, error) {
	users, err := service.userRepository.GetAllUserSecret()
	return users, err
}
