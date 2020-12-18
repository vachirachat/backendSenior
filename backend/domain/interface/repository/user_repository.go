package repository

import (
	"backendSenior/domain/model"
)

type UserRepository interface {
	GetAllUser() ([]model.User, error)
	GetLastUser() (model.User, error)
	GetAllUserToken() ([]model.UserToken, error)

	GetUserByID(userID string) (model.User, error)
	AddUser(user model.User) error
	EditUserName(userID string, user model.User) error
	EditUserRole(model.UserSecret) error
	DeleteUserByID(userID string) error
	GetUserByEmail(email string) (model.User, error)

	//login
	GetUserTokenById(email string) (model.UserToken, error)
	GetUserIdByToken(token string) (model.UserToken, error)
	GetUserRole(userID string) (string, error)
	AddToken(UserToken model.UserToken) error
	GetUserSecret(userSecret model.UserSecret) (model.User, error)
	//SignUp
	AddUserSecrect(user model.UserSecret) error
	GetAllUserSecret() ([]model.UserSecret, error)
	//GetRoomWithRoomID(roomID bson.ObjectId) (model.Room, error)
}
