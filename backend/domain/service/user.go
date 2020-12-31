package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/service/auth"
	"errors"
	"log"

	"backendSenior/domain/model"
	"backendSenior/utills"

	"github.com/globalsign/mgo/bson"
	"golang.org/x/crypto/bcrypt"
)

// UserService provide access to user related functions
type UserService struct {
	userRepository repository.UserRepository
	jwtService     *auth.JWTService
}

// NewUserService return instance of user service
func NewUserService(userRepo repository.UserRepository, jwtService *auth.JWTService) *UserService {
	return &UserService{
		userRepository: userRepo,
		jwtService:     jwtService,
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

// GetUsersByIDs return multiple user
func (service *UserService) GetUsersByIDs(userIDs []string) ([]model.User, error) {
	users, err := service.userRepository.GetUsersByIDs(userIDs)
	return users, err
}

// GetUserByEmail return user with specified email
func (service *UserService) GetUserByEmail(email string) (model.User, error) {
	user, err := service.userRepository.GetUserByEmail(email)
	return user, err
}

// AddUser create a user
func (service *UserService) AddUser(user model.User) error {

	err := service.userRepository.AddUser(user)
	return err
}

// UpdateUser update whole user
func (service *UserService) UpdateUser(userID string, user model.User) error {
	err := service.userRepository.UpdateUser(userID, user)
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

func (service *UserService) Login(email string, password string) (model.User, error) {
	// var token model.TokenDetails
	user, err := service.GetUserByEmail(email)
	log.Println(user)
	if err != nil {
		return model.User{}, errors.New("User not exists")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

// //Login find user with matching username, password, isAdmin, return token
// func (service *UserService) EditUserRole(credentials model.UserSecret) error {
// 	_, err := service.userRepository.GetUserSecret(credentials)
// 	if err != nil {
// 		return errors.New("User not exists")
// 	}

// 	exist, _ := utills.In_array(strings.ToLower(credentials.Role), []string{utills.ROLEADMIN, utills.ROLEUSER})
// 	if !exist {
// 		return errors.New("Invalid Role")
// 	}
// 	err = service.userRepository.EditUserRole(credentials)
// 	if err != nil {
// 		return errors.New("Cannot Update exists")
// 	}

// 	return nil
// }

// Signup API
func (service *UserService) Signup(user model.User) error {
	_, err := service.userRepository.GetUserByEmail(user.Email)
	if err == nil {
		return errors.New("User already exists")
	}

	// Add User to DB
	user.Password = utills.HashPassword(user.Password)
	user.UserType = "user"
	user.Room=     []bson.ObjectId{}
	user.Organize= []bson.ObjectId{}
	user.FCMTokens = []string{}
	err = service.userRepository.AddUser(user)
	if err != nil {
		return err
	}

	return nil
}
