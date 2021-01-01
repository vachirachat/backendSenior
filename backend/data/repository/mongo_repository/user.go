package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"log"
	"unsafe"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"golang.org/x/crypto/bcrypt"
)

type UserRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

var _ repository.UserRepository = (*UserRepositoryMongo)(nil)

const (
	collectionToken  = "UserToken"
	collectionSecret = "UserSecret"
)

func (userMongo UserRepositoryMongo) GetAllUser() ([]model.User, error) {
	var Users []model.User
	err := userMongo.ConnectionDB.DB(dbName).C(collectionUser).Find(nil).All(&Users)
	return Users, err
}

func (userMongo UserRepositoryMongo) GetAllUserSecret() ([]model.UserSecret, error) {
	var Users []model.UserSecret
	err := userMongo.ConnectionDB.DB(dbName).C(collectionSecret).Find(nil).All(&Users)
	return Users, err
}

func (userMongo UserRepositoryMongo) GetUserByID(userID string) (model.User, error) {
	var user model.User
	objectID := bson.ObjectIdHex(userID)
	err := userMongo.ConnectionDB.DB(dbName).C(collectionUser).FindId(objectID).One(&user)
	return user, err
}

// GetUsersByIDs query users by array of IDs
func (userMongo UserRepositoryMongo) GetUsersByIDs(userID []string) ([]model.User, error) {
	var users []model.User
	err := userMongo.ConnectionDB.DB(dbName).C(collectionUser).Find(idInArr(userID)).All(&users)
	return users, err
}

func (userMongo UserRepositoryMongo) GetLastUser() (model.User, error) {
	var user model.User
	err := userMongo.ConnectionDB.DB(dbName).C(collectionUser).Find(nil).Sort("-created_time").One(&user)
	return user, err
}

func (userMongo UserRepositoryMongo) AddUser(user model.User) error {
	userInsert := *(*model.UserInsert)(unsafe.Pointer(&user))

	return userMongo.ConnectionDB.DB(dbName).C(collectionUser).Insert(userInsert)
}

func (userMongo UserRepositoryMongo) UpdateUser(userID string, user model.User) error {
	objectID := bson.ObjectIdHex(userID)
	// dont allow update these fields
	user.UserID = ""
	return userMongo.ConnectionDB.DB(dbName).C(collectionUser).UpdateId(objectID, bson.M{"$set": user})
}

func (userMongo UserRepositoryMongo) DeleteUserByID(userID string) error {
	objectID := bson.ObjectIdHex(userID)
	return userMongo.ConnectionDB.DB(dbName).C(collectionUser).RemoveId(objectID)
}

//  Token
func (userMongo UserRepositoryMongo) AddToken(UserToken model.UserToken) error {
	return userMongo.ConnectionDB.DB(dbName).C(collectionToken).Insert(UserToken)
}

type UserSecret struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
	Role     string `json:"role" bson:"role"`
}

func (userMongo UserRepositoryMongo) EditUserRole(userSecret model.UserSecret) error {
	mapSecret := bson.M{"email": userSecret.Email}
	newName := bson.M{"$set": bson.M{"role": userSecret.Role}}
	return userMongo.ConnectionDB.DB(dbName).C(collectionUser).Update(mapSecret, newName)
}

func (userMongo UserRepositoryMongo) GetUserTokenById(userID string) (model.UserToken, error) {
	var userToken model.UserToken
	err := userMongo.ConnectionDB.DB(dbName).C(collectionToken).FindId(bson.ObjectIdHex(userID)).One(&userToken)
	return userToken, err
}

func (userMongo UserRepositoryMongo) GetUserRole(userID string) (string, error) {
	var userRole model.UserSecret
	err := userMongo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.ObjectIdHex(userID)).One(&userRole)
	return userRole.Role, err
}

func (userMongo UserRepositoryMongo) GetUserIdByToken(token string) (model.UserToken, error) {
	var userToken model.UserToken
	err := userMongo.ConnectionDB.DB(dbName).C(collectionToken).Find(bson.M{"token": token}).One(&userToken)
	return userToken, err
}

func (userMongo UserRepositoryMongo) GetAllUserToken() ([]model.UserToken, error) {
	var UsersToken []model.UserToken
	err := userMongo.ConnectionDB.DB(dbName).C(collectionToken).Find(nil).All(&UsersToken)
	return UsersToken, err
}

func (userMongo UserRepositoryMongo) GetUserSecret(userCredential model.UserSecret) (model.User, error) {
	var userCred model.UserSecret
	var user model.User
	err := userMongo.ConnectionDB.DB(dbName).C(collectionSecret).Find(bson.M{"email": userCredential.Email}).One(&userCred)
	if err != nil {
		log.Println("User dose not exist")
		return user, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(userCred.Password), []byte(userCredential.Password))
	if err != nil {
		log.Println("Password dose not exist")
		return user, err
	}
	err = userMongo.ConnectionDB.DB(dbName).C(collectionUser).Find(bson.M{"email": userCred.Email}).One(&user)
	if err != nil {
		return user, err
	}
	return user, err

}

func (userMongo UserRepositoryMongo) GetUserByEmail(email string) (model.User, error) {
	var user model.User

	err := userMongo.ConnectionDB.DB(dbName).C(collectionUser).Find(bson.M{"email": email}).One(&user)
	return user, err
}

// user secrect
func (userMongo UserRepositoryMongo) AddUserSecrect(user model.UserSecret) error {
	return userMongo.ConnectionDB.DB(dbName).C(collectionSecret).Insert(user)
}

func (userMongo UserRepositoryMongo) GetUserRoomByUserID(userID string) ([]string, error) {
	var user model.User
	err := userMongo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.ObjectIdHex(userID)).One(&user)
	return model.ToStringArr(user.Room), err
}
