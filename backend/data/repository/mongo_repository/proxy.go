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

func orEmptyProxy(slice []model.Proxy) []model.Proxy {
	if slice == nil {
		return make([]model.Proxy, 0)
	}
	return slice
}

// AddProxy create new proxy and return ID
func (repo *ProxyRepositoryMongo) AddProxy(proxy model.Proxy) (string, error) {
	proxyID := bson.NewObjectId()
	proxy.ProxyID = proxyID
	proxy.Rooms = []bson.ObjectId{}

	err := repo.conn.DB(dbName).C(collectionProxy).Insert(proxy)
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
	return orEmptyProxy(proxies), nil
}
func (repo *ProxyRepositoryMongo) DeleteProxy(proxyID string) error {
	err := repo.conn.DB(dbName).C(collectionProxy).RemoveId(bson.ObjectIdHex(proxyID))
	return err
}
func (repo *ProxyRepositoryMongo) UpdateProxy(proxyID string, update model.Proxy) error {
	update.ProxyID = ""
	err := repo.conn.DB(dbName).C(collectionProxy).UpdateId(bson.ObjectIdHex(proxyID), bson.M{"$set": update})
	return err
}
func (repo *ProxyRepositoryMongo) GetByID(proxyID string) (model.Proxy, error) {
	var proxy model.Proxy
	err := repo.conn.DB(dbName).C(collectionProxy).FindId(bson.ObjectIdHex(proxyID)).One(&proxy)
	return proxy, err
}

// GetByIDs return multiple proxies by specifying array of IDs
func (repo *ProxyRepositoryMongo) GetByIDs(proxyIDs []string) ([]model.Proxy, error) {
	var proxies []model.Proxy
	err := repo.conn.DB(dbName).C(collectionProxy).Find(idInArr(proxyIDs)).All(&proxies)
	return proxies, err
}

func (repo *ProxyRepositoryMongo) GetByRoom(roomID string) ([]model.Proxy, error) {
	var proxies []model.Proxy
	err := repo.conn.DB(dbName).C(collectionProxy).Find(bson.M{
		"rooms": bson.ObjectIdHex(roomID),
	}).All(&proxies)
	return proxies, err
}
