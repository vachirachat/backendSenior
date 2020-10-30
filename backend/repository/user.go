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

	//login
	GetUserTokenById(email string) (model.UserToken, error)
	GetUserIdByToken(token string) (model.UserToken, error)
	GetAllUserToken() ([]model.UserToken, error)

	AddToken(UserToken model.UserToken) error
	GetUserLogin(userLogin model.UserLogin) (model.UserLogin, error)
	//SignUp
	AddUserSecrect(user model.UserLogin) error
}

type UserRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

const (
	DBNameUser       = "User"
	collectionUser   = "UserData"
	collectionToken  = "UserToken"
	collectionSecret = "UserSecret"
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

//  Token
func (userMongo UserRepositoryMongo) AddToken(UserToken model.UserToken) error {
	return userMongo.ConnectionDB.DB(DBNameUser).C(collectionToken).Insert(UserToken)
}

// Implemet more
// func (userMongo UserRepositoryMongo) EditToken(UserToken model.UserToken) error {
// 	newToken := bson.M{"$set": bson.M{"Token": UserToken.Token}}
// 	return userMongo.ConnectionDB.DB(DBNameUser).C(collectionToken).Update(bson.M{"Email": Email}, newToken)
// }

func (userMongo UserRepositoryMongo) GetUserTokenById(Email string) (model.UserToken, error) {
	var userToken model.UserToken
	//objectID := bson.ObjectIdHex(userID)
	err := userMongo.ConnectionDB.DB(DBNameUser).C(collectionToken).Find(bson.M{"Email": Email}).One(&userToken)
	return userToken, err
}

func (userMongo UserRepositoryMongo) GetUserIdByToken(token string) (model.UserToken, error) {
	var userToken model.UserToken
	//objectID := bson.ObjectIdHex(userID)
	err := userMongo.ConnectionDB.DB(DBNameUser).C(collectionToken).Find(bson.M{"Token": token}).One(&userToken)
	return userToken, err
}

func (userMongo UserRepositoryMongo) GetAllUserToken() ([]model.UserToken, error) {
	var UsersToken []model.UserToken
	err := userMongo.ConnectionDB.DB(DBNameUser).C(collectionToken).Find(nil).All(&UsersToken)
	return UsersToken, err
}

func (userMongo UserRepositoryMongo) GetUserLogin(userLogin model.UserLogin) (model.UserLogin, error) {
	var user model.UserLogin
	err := userMongo.ConnectionDB.DB(DBNameUser).C(collectionSecret).Find(bson.M{"username": userLogin.Username, "password": userLogin.Password}).One(&user)
	return user, err
}

// user secrect
func (userMongo UserRepositoryMongo) AddUserSecrect(user model.UserLogin) error {
	return userMongo.ConnectionDB.DB(DBNameUser).C(collectionSecret).Insert(user)
}

// oauth Add token
func AddToken(UserToken model.UserToken) error {
	var ConnectionDB *mgo.Session
	return ConnectionDB.DB(DBNameUser).C(collectionToken).Insert(UserToken)
}
