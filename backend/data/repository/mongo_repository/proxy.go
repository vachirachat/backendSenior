package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type ProxyRepositoryMongo struct {
	conn *mgo.Session
}

func NewProxyRepositoryMongo(conn *mgo.Session) *ProxyRepositoryMongo {
	return &ProxyRepositoryMongo{
		conn: conn,
	}
}

var _ repository.ProxyRepository = (*ProxyRepositoryMongo)(nil)

func orEmpty(slice []model.Proxy) []model.Proxy {
	if slice == nil {
		return make([]model.Proxy, 0)
	}
	return slice
}

// AddProxy create new proxy and return ID
func (repo *ProxyRepositoryMongo) AddProxy(name string) (string, error) {
	proxyID := bson.NewObjectId()
	err := repo.conn.DB(dbName).C(collectionProxy).Insert(model.Proxy{
		ProxyID: proxyID,
		Name:    name,
		Rooms:   []bson.ObjectId{},
	})
	if err != nil {
		return "", err
	}
	return proxyID.Hex(), nil
}

// GetAllProxies returns all proxies
func (repo *ProxyRepositoryMongo) GetAllProxies() ([]model.Proxy, error) {
	var proxies []model.Proxy
	err := repo.conn.DB(dbName).C(collectionProxy).Find(nil).All(&proxies)
	if err != nil {
		return nil, err
	}
	return orEmpty(proxies), nil
}
func (repo *ProxyRepositoryMongo) DeleteProxy(proxyID string) error {
	panic("Not implemented")

}
func (repo *ProxyRepositoryMongo) UpdateProxy(proxyID string, update model.Proxy) error {
	panic("Not implemented")

}
func (repo *ProxyRepositoryMongo) GetByID(proxyID string) (model.Proxy, error) {
	panic("Not implemented")
}
