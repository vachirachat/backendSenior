package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// FCMUserRepository manage relation between user and token
type FCMUserRepository struct {
	conn *mgo.Session
}

// NewFCMUserRepository create new instance of FCMUserRepository
func NewFCMUserRepository(conn *mgo.Session) *FCMUserRepository {
	return &FCMUserRepository{
		conn: conn,
	}
}

var _ repository.FCMUserRepository = (*FCMUserRepository)(nil)

func orEmptyStr(slice []string) []string {
	if slice == nil {
		return make([]string, 0)
	}
	return slice
}

// GetUserTokens returns array of user token
func (repo *FCMUserRepository) GetUserTokens(userID string) ([]string, error) {
	var user model.User
	err := repo.conn.DB(dbName).C(collectionUser).FindId(bson.ObjectIdHex(userID)).One(&user)
	if err != nil {
		return []string{}, err
	}

	return orEmptyStr(user.FCMTokens), nil
}

// GetTokenOwner return userId of token owner
func (repo *FCMUserRepository) GetTokenOwner(token string) (string, error) {
	var foundToken model.FCMToken
	err := repo.conn.DB(dbName).C(collectionFCMToken).FindId(token).One(&foundToken)
	if err != nil {
		return "", err
	}

	return foundToken.UserID.Hex(), nil
}

// AddUserToken add token to user's token list
// Note that it DOES NOT create user token
func (repo *FCMUserRepository) AddUserToken(userID string, tokenID string) error {
	err := repo.conn.DB(dbName).C(collectionUser).UpdateId(bson.ObjectIdHex(userID), bson.M{
		"$addToSet": bson.M{
			"fcmTokens": tokenID,
		},
	})
	return err
}

// DeleteUserToken remove token from user's token list
// Note that it DOES NOT delete user token
func (repo *FCMUserRepository) DeleteUserToken(userID string, tokenID string) error {
	err := repo.conn.DB(dbName).C(collectionUser).UpdateId(bson.ObjectIdHex(userID), bson.M{
		"$pull": bson.M{
			"fcmTokens": tokenID,
		},
	})
	return err
}
