package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type OrganizeRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

func NewOrganizeRepositoryMongo(conn *mgo.Session) *OrganizeRepositoryMongo {
	return &OrganizeRepositoryMongo{
		ConnectionDB: conn,
	}
}

var _ repository.OrganizeRepository = (*OrganizeRepositoryMongo)(nil)

func (organizeMongo *OrganizeRepositoryMongo) GetAllOrganize() ([]model.Organize, error) {
	var organizes []model.Organize
	err := organizeMongo.ConnectionDB.DB(dbName).C(collectionOrganize).Find(nil).All(&organizes)
	return organizes, err
}

// GetOrganizeByUser return all organization user is in
func (organizeMongo *OrganizeRepositoryMongo) GetOrganizeByUser(userID string) ([]model.Organize, error) {
	var organizes []model.Organize
	err := organizeMongo.ConnectionDB.DB(dbName).C(collectionOrganize).Find(bson.M{
		"members": bson.ObjectIdHex(userID),
	}).All(&organizes)
	return organizes, err
}

func (organizeMongo *OrganizeRepositoryMongo) CreateOrganize(organize model.Organize) (string, error) {
	organize.OrganizeID = bson.NewObjectId()
	cnt, err := organizeMongo.ConnectionDB.DB(dbName).C(collectionOrganize).FindId(organize.OrganizeID).Count()
	if cnt == 1 || err != nil {
		return "", fmt.Errorf("room error: %s", err)
	}
	err = organizeMongo.ConnectionDB.DB(dbName).C(collectionOrganize).Insert(organize)
	return organize.OrganizeID.Hex(), err
}

func (organizeMongo OrganizeRepositoryMongo) DeleteOrganize(organizeID string) error {
	objectID := bson.ObjectIdHex(organizeID)
	return organizeMongo.ConnectionDB.DB(dbName).C(collectionOrganize).RemoveId(objectID)
}

func (organizeMongo OrganizeRepositoryMongo) GetOrganizeById(organizeID string) (model.Organize, error) {
	var organize model.Organize
	err := organizeMongo.ConnectionDB.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(organizeID)).One(&organize)
	return organize, err
}

// GetOrganizesByIds query multiple organizations by array of IDs
func (organizeMongo OrganizeRepositoryMongo) GetOrganizesByIDs(organizeIDs []string) ([]model.Organize, error) {
	var orgs []model.Organize
	err := organizeMongo.ConnectionDB.DB(dbName).C(collectionOrganize).Find(idInArr(organizeIDs)).All(&orgs)
	return orgs, err
}

func (organizeMongo OrganizeRepositoryMongo) UpdateOrganize(organizeID string, name string) error {
	return organizeMongo.ConnectionDB.DB(dbName).C(collectionOrganize).UpdateId(bson.ObjectIdHex(organizeID), bson.M{
		"$set": bson.M{"name": name},
	})
}
