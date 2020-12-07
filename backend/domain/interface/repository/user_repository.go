package repository

import (
	"backendSenior/domain/model"

	"github.com/globalsign/mgo/bson"
)

type UserRepository interface {
	GetAllUser() ([]model.User, error)
	GetLastUser() (model.User, error)
	GetUserByID(userID bson.ObjectId) (model.User, error)
	AddUser(user model.User) error
	EditUserName(userID bson.ObjectId, user model.User) error
	DeleteUserByID(userID string) error
	GetUserByEmail(email string) (model.User, error)

	//login
	GetUserTokenById(email string) (model.UserToken, error)
	GetUserIdByToken(token string) (model.UserToken, error)
	GetAllUserToken() ([]model.UserToken, error)

	AddToken(UserToken model.UserToken) error
	GetUserLogin(userLogin model.UserLogin) (model.UserLogin, error)
	//SignUp
	AddUserSecrect(user model.UserLogin) error

	GetAllUserSecret() ([]model.UserLogin, error)
	//GetRoomWithRoomID(roomID bson.ObjectId) (model.Room, error)
}
