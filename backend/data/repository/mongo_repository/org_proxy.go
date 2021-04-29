package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/utills"
	"errors"
	"fmt"
	"log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/globalsign/mgo/txn"
)

type OrgProxyRepository struct {
	conn      *mgo.Session
	txnRunner *txn.Runner
}

func NewOrgProxyRepositoryMongo(conn *mgo.Session) *OrgProxyRepository {
	runner := txn.NewRunner(conn.DB(dbName).C(collectionTXNRoomUser))
	return &OrgProxyRepository{
		conn:      conn,
		txnRunner: runner,
	}
}

var _ repository.OrgProxyRepository = (*OrgProxyRepository)(nil)

// GetOrgRooms return proxiesID of org
func (repo *OrgProxyRepository) GetOrgProxyIDs(orgID string) (proxiseIDs []model.Proxy, err error) {
	var org model.Organize
	err = repo.conn.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(orgID)).One(&org)
	if err != nil {
		return nil, err
	}
	// Fix can do more effective
	proxies := []model.Proxy{}
	for _, v := range org.Proxies {
		var proxy model.Proxy
		err = repo.conn.DB(dbName).C(collectionProxy).FindId(v).One(&proxy)
		if err != nil {
			return nil, err
		}
		proxies = append(proxies, proxy)
	}
	// Fix can do more effective
	return proxies, nil
}

// AddProxiseToOrg adds proxies to the org, each proxy must don't have any org
func (repo *OrgProxyRepository) AddProxiseToOrg(orgID string, proxyIDs []string) (err error) {
	n, err := repo.conn.DB(dbName).C(collectionProxy).Find(idInArr(proxyIDs)).Count()

	if err != nil {
		return err
	}
	if n != len(proxyIDs) {
		return fmt.Errorf("Invalid ProxyIDs, some of them not exists %d/%d", n, len(proxyIDs))
	}

	n, err = repo.conn.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(orgID)).Count()
	if err != nil || n != 1 {
		return errors.New("Invalid orgID ID")
	}

	ops := []txn.Op{
		{
			C:  collectionOrganize,
			Id: bson.ObjectIdHex(orgID),
			Update: bson.M{
				"$addToSet": model.OrganizationT{
					Proxies: bson.M{
						"$each": utills.ToObjectIdArr(proxyIDs),
					},
				},
			},
		},
	}

	log.Println("proxyIDs in repo >>", proxyIDs)
	log.Println("OrgID in repo >>", orgID)
	for _, proxy := range proxyIDs {
		ops = append(ops, txn.Op{
			C:  collectionProxy,
			Id: bson.ObjectIdHex(proxy),
			// Assert: model.ProxyUpdateMongo{ // assert orgId doesn't exist
			// 	Org: bson.M{
			// 		"$eq": nil,
			// 	},
			// },
			Update: bson.M{
				"$set": model.ProxyUpdateMongo{
					Org: bson.ObjectIdHex(orgID),
				},
			},
		})
	}

	err = repo.txnRunner.Run(ops, "", nil)
	if err != nil {
		return err
	}

	return nil
}

// RemoveProxiseFromOrg delete proxies to the org
func (repo *OrgProxyRepository) RemoveProxiseFromOrg(orgID string, proxyIDs []string) (err error) {
	n, err := repo.conn.DB(dbName).C(collectionProxy).Find(idInArr(proxyIDs)).Count()

	if err != nil {
		return err
	}

	if n != len(proxyIDs) {
		return fmt.Errorf("Invalid ProxyIDs, some of them not exists %d/%d", n, len(proxyIDs))
	}

	n, err = repo.conn.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(orgID)).Count()
	if err != nil || n != 1 {
		return errors.New("Invalid orgID ID")
	}

	ops := []txn.Op{
		{
			C:  collectionOrganize,
			Id: bson.ObjectIdHex(orgID),
			Update: bson.M{
				"$pullAll": model.OrganizationT{
					Proxies: utills.ToObjectIdArr(proxyIDs),
				},
			},
		},
	}
	for _, proxyID := range proxyIDs {
		ops = append(ops, txn.Op{
			C:  collectionProxy,
			Id: bson.ObjectIdHex(proxyID),
			Update: bson.M{
				"$unset": model.ProxyUpdateMongo{
					Org: "", // value doesn't matter
				},
			},
		})
	}

	err = repo.txnRunner.Run(ops, "", nil)
	if err != nil {
		return err
	}

	return nil
}
