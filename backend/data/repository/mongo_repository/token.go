package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"github.com/globalsign/mgo"
)

type TokenRepository struct {
	conn *mgo.Session
	col  *mgo.Collection // shortcut to access token collection
}

// NewFCMUserRepository create new instance of FCMUserRepository
func NewTokenRepository(conn *mgo.Session) *TokenRepository {
	return &TokenRepository{
		conn: conn,
		col:  conn.DB(dbName).C(collectionToken),
	}
}

var _ repository.TokenRepository = (*TokenRepository)(nil)

func (repo *TokenRepository) CountToken(filter interface{}) (int, error) {
	cnt, err := repo.col.Find(filter).Count()
	return cnt, err
}

func (repo *TokenRepository) InsertToken(token model.TokenDB) error {

	err := repo.col.Insert(token)
	if err != nil {
		return err
	}
	return nil
}

func (repo *TokenRepository) RemoveTokens(filter interface{}) (int, error) {
	info, err := repo.col.RemoveAll(filter)
	if err != nil {
		return 0, err
	}
	return info.Removed, nil
}
