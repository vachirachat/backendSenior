package repository

import (
	"backendSenior/model"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type UserRepository interface {
	GetAllUser() ([]model.User, error)
	GetLastUser() (model.User, error)
	GetUserByID(userID string) (model.User, error)
	AddUser(user model.User) error
	EditUserName(userID string, user model.User) error
	DeleteUserByID(userID string) error
}

type UserRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

const (
	DBNameUser     = "User"
	collectionUser = "UserData"
)

func (userMongo UserRepositoryMongo) GetAllUser() ([]model.User, error) {
	var Users []model.User
	err := userMongo.ConnectionDB.DB(DBNameUser).C(collectionUser).Find(nil).All(&Users)
	return Users, err
}

func (userMongo UserRepositoryMongo) GetUserByID(userID string) (model.User, error) {
	var user model.User
	objectID := bson.ObjectIdHex(userID)
	err := userMongo.ConnectionDB.DB(DBNameUser).C(collectionUser).FindId(objectID).One(&user)
	return user, err
}
func (userMongo UserRepositoryMongo) GetLastUser() (model.User, error) {
	var user model.User
	err := userMongo.ConnectionDB.DB(DBNameUser).C(collectionUser).Find(nil).Sort("-created_time").One(&user)
	return user, err
}
func (userMongo UserRepositoryMongo) AddUser(user model.User) error {
	return userMongo.ConnectionDB.DB(DBNameUser).C(collectionUser).Insert(user)
}

func (userMongo UserRepositoryMongo) EditUserName(userID string, user model.User) error {
	objectID := bson.ObjectIdHex(userID)
	newName := bson.M{"$set": bson.M{"user_name": user.Name, "updated_time": time.Now()}}
	return userMongo.ConnectionDB.DB(DBNameUser).C(collectionUser).UpdateId(objectID, newName)
}

func (userMongo UserRepositoryMongo) DeleteUserByID(userID string) error {
	objectID := bson.ObjectIdHex(userID)
	return userMongo.ConnectionDB.DB(DBNameUser).C(collectionUser).RemoveId(objectID)
}
