package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/service/auth"
	"errors"
	"strings"

	"backendSenior/domain/model"
	"backendSenior/utills"
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

func (service *UserService) GetUserIdByToken(token string) (model.UserToken, error) {
	userToken, err := service.userRepository.GetUserIdByToken(token)
	return userToken, err
}

func (service *UserService) GetUserByID(userID string) (model.User, error) {
	user, err := service.userRepository.GetUserByID(userID)
	return user, err
}

// for get user by id
// func (api UserAPI) GetUserByID(context *gin.Context) {
// 	userID := context.Param("user_id")
// 	// user, err := service.userRepository.GetUserByID(userID)
// 	if err != nil {
// 		log.Println("error GetUserByID", err.Error())
// 		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
// 		return
// 	}
// 	context.JSON(http.StatusOK, user)
// }

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
func (service *UserService) GetUserRole(userID string) (string, error) {
	token, err := service.userRepository.GetUserRole(userID)
	return token, err
}

type messageLogin struct {
	status string
	Email  string
	token  string
}

//Login find user with matching username, password, isAdmin, return token
// func (service *UserService) Login(credentials model.UserSecret) (string, string, error) {
// 	user, err := service.userRepository.GetUserSecret(credentials)
// 	if err != nil {
// 		return "", "", errors.New("User not exists")
// 	}
// 	var usertoken model.UserToken
// 	usertoken, err = service.userRepository.GetUserTokenById(user.UserID.Hex())

// 	timeExp, _ := time.Parse(time.RFC3339, usertoken.TimeExpired)

// 	if err != nil || !timeExp.Before(time.Now()) {
// 		usertoken.UserID = user.UserID
// 		usertoken.Token = ksuid.New().String()
// 		usertoken.TimeExpired = time.Now().Add(24 * time.Hour).String()
// 		err = service.userRepository.AddToken(usertoken)
// 	}

// 	return usertoken.Token, usertoken.TimeExpired, nil
// }

func (service *UserService) Login(credentials model.UserSecret) (*model.TokenDetails, error) {
	// var token model.TokenDetails
	user, err := service.userRepository.GetUserSecret(credentials)
	if err != nil {
		return nil, errors.New("User not exists")
	}
	userId := user.UserID.Hex()
	token, err := auth.CreateToken(userId)
	if err != nil {
		return nil, errors.New("Create Token Fail")
	}
	return token, nil
}

//Login find user with matching username, password, isAdmin, return token
func (service *UserService) EditUserRole(credentials model.UserSecret) error {
	_, err := service.userRepository.GetUserSecret(credentials)
	if err != nil {
		return errors.New("User not exists")
	}

	exist, _ := utills.In_array(strings.ToLower(credentials.Role), []string{utills.ROLEADMIN, utills.ROLEUSER})
	if !exist {
		return errors.New("Invalid Role")
	}
	err = service.userRepository.EditUserRole(credentials)
	if err != nil {
		return errors.New("Cannot Update exists")
	}

	return nil
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

	err = service.userRepository.AddUserSecrect(model.UserSecret{
		Email:    user.Email,
		Password: user.Password,
		Role:     "user",
	})
	if err != nil {
		return err
	}

	return nil
}

//GetAllUserSecret return secret of all user
func (service *UserService) GetAllUserSecret() ([]model.UserSecret, error) {
	users, err := service.userRepository.GetAllUserSecret()
	return users, err
}
