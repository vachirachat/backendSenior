package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"unsafe"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type TokenRepository struct {
	conn *mgo.Session
}

// NewFCMUserRepository create new instance of FCMUserRepository
func NewTokenRepository(conn *mgo.Session) *TokenRepository {
	return &TokenRepository{
		conn: conn,
	}
}

var _ repository.TokenRepository = (*TokenRepository)(nil)

func (repo *TokenRepository) GetAllToken() ([]model.TokenDB, error) {
	var tokens []model.TokenDB
	err := repo.conn.DB(dbName).C(collectionToken).Find(nil).All(&tokens)
	return tokens, err
}

func (repo *TokenRepository) VerifyDBToken(userid string, accessToken string) (string, error) {
	var token model.TokenDB
	err := repo.conn.DB(dbName).C(collectionToken).FindId(bson.ObjectIdHex(userid)).One(&token)
	return token.AccessToken, err
}

func (repo *TokenRepository) AddToken(userid string, accessToken string) error {
	var token model.TokenDB
	token.AccessToken = accessToken
	token.UserID = bson.ObjectIdHex(userid)
	tokenInsert := *(*model.TokenDBInsert)(unsafe.Pointer(&token))
	err := repo.conn.DB(dbName).C(collectionToken).Insert(tokenInsert)
	if err != nil {
		update := map[string]interface{}{
			"accesstoken": accessToken,
		}
		// err = repo.conn.DB(dbName).C(collectionToken).UpdateId(bson.ObjectIdHex(userid), bson.M{"$set": update})
		err = repo.conn.DB(dbName).C(collectionToken).Update(bson.M{"_id": bson.ObjectIdHex(userid)}, bson.M{"$set": update})
	}
	return err
}

func (repo *TokenRepository) RemoveToken(userid string) error {
	update := map[string]interface{}{
		"accesstoken": "",
	}
	return repo.conn.DB(dbName).C(collectionToken).UpdateId(bson.ObjectIdHex(userid), bson.M{"$set": update})
}
