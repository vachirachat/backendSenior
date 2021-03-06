package repository

import (
	"backendSenior/domain/model"

	"github.com/globalsign/mgo/bson"
)

type UserRepository interface {
	GetAllUser() ([]model.User, error)
	GetLastUser() (model.User, error)
	GetAllUserToken() ([]model.UserToken, error)

	GetUserByID(userID string) (model.User, error)
	GetUsersByIDs(userIDs []string) ([]model.User, error)
	AddUser(user model.User) error
	UpdateUser(userID string, user model.User) error
	EditUserRole(model.User) error
	DeleteUserByID(userID string) error
	GetUserByEmail(email string) (model.User, error)

	BulkUpdateUser([]bson.ObjectId, model.UserUpdateMongo) error

	GetUserRole(userID string) (string, error)
	// GetUserSecret(userSecret model.UserSecret) (model.User, error)
}
