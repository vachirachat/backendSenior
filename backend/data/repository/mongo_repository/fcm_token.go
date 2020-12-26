package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// FCMTokenRepository manage FCM Tokens, storing them in mongoDB
type FCMTokenRepository struct {
	conn *mgo.Session
}

// NewFCMTokenRepository create new FCMTokenRepository instance
func NewFCMTokenRepository(conn *mgo.Session) *FCMTokenRepository {
	return &FCMTokenRepository{
		conn: conn,
	}
}

var _ repository.FCMTokenRepository = (*FCMTokenRepository)(nil)

// GetAllTokens return all tokens
func (repo *FCMTokenRepository) GetAllTokens() ([]model.FCMToken, error) {
	var fcmtokens []model.FCMToken
	err := repo.conn.DB(dbName).C(collectionFCMToken).Find(nil).All(&fcmtokens)
	return fcmtokens, err
}

// GetTokenByID return token by ID (id is token itself)
func (repo *FCMTokenRepository) GetTokenByID(token string) (model.FCMToken, error) {
	var fcmtoken model.FCMToken
	err := repo.conn.DB(dbName).C(collectionFCMToken).FindId(token).One(&fcmtoken)
	return fcmtoken, err
}

// GetTokensByIDs return multiple tokens by array of ID
func (repo *FCMTokenRepository) GetTokensByIDs(tokens []string) ([]model.FCMToken, error) {
	var fcmtokens []model.FCMToken
	err := repo.conn.DB(dbName).C(collectionFCMToken).FindId(bson.M{"$in": tokens}).All(&fcmtokens)
	return fcmtokens, err
}

// AddToken insert token into database
// it would error if it already exists
func (repo *FCMTokenRepository) AddToken(token model.FCMToken) error {
	err := repo.conn.DB(dbName).C(collectionFCMToken).Insert(token)
	return err
}

// DeleteToken delete token
func (repo *FCMTokenRepository) DeleteToken(token string) error {
	return repo.conn.DB(dbName).C(collectionFCMToken).RemoveId(token)
}

// UpdateToken update token details
func (repo *FCMTokenRepository) UpdateToken(token string, update model.FCMToken) error {
	return repo.conn.DB(dbName).C(collectionFCMToken).UpdateId(token, bson.M{
		"$set": update,
	})
}
