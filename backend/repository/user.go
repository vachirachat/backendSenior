package repository

import (
	"backendSenior/model"
	"backendSenior/utills"

	"github.com/globalsign/mgo"
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

func (userMongo UserRepositoryMongo) GetUserByID(userID bson.ObjectId) (model.User, error) {
	var user model.User
	// objectID := bson.ObjectIdHex(userID)
	err := userMongo.ConnectionDB.DB(DBNameUser).C(collectionUser).FindId(userID).One(&user)
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

func (userMongo UserRepositoryMongo) EditUserName(userID bson.ObjectId, user model.User) error {
	// objectID := bson.ObjectIdHex(userID)
	newName := bson.M{"$set": bson.M{"name": user.Name, "email": user.Email, "password": user.Password, "room": user.Room, "userType": user.UserType}}
	return userMongo.ConnectionDB.DB(DBNameUser).C(collectionUser).UpdateId(userID, newName)
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
	err := userMongo.ConnectionDB.DB(DBNameUser).C(collectionSecret).Find(bson.M{"username": userLogin.Username, "password": utills.HashPassword(userLogin.Password)}).One(&user)
	return user, err
}

func (userMongo UserRepositoryMongo) GetUserByEmail(email string) (model.User, error) {
	var user model.User
	err := userMongo.ConnectionDB.DB(DBNameUser).C(collectionUser).Find(bson.M{"email": email}).One(&user)
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

func GetUserIdByToken(token string) (model.UserToken, error) {
	var userToken model.UserToken
	var ConnectionDB *mgo.Session
	//objectID := bson.ObjectIdHex(userID)
	err := ConnectionDB.DB(DBNameUser).C(collectionToken).Find(bson.M{"Token": token}).One(&userToken)
	return userToken, err
}
