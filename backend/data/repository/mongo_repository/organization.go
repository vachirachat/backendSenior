package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"

	// "fmt"
	"log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type OrganizationRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

var _ repository.OrganizationRepository = (*OrganizationRepositoryMongo)(nil)

// finish
// GetAllMessages return all message from all rooms with optional time filter
func (organizationMongo *OrganizationRepositoryMongo) GetAllOrganization() ([]model.Organization, error) {
	var organizations []model.Organization
	err := organizationMongo.ConnectionDB.DB(dbName).C(collectionOrganization).Find(nil).All(&organizations)
	return organizations, err
}

//finish
// GetMessagesByRoom return messages from specified room, with optional time filter
func (organizationMongo *OrganizationRepositoryMongo) GetMemberInOrganization(orgID string) ([]model.User, error) {
	var organization model.Organization
	err := organizationMongo.ConnectionDB.DB(dbName).C(collectionOrganization).Find(bson.ObjectIdHex(orgID)).All(&organization)
	log.Println(organization)
	var userList []model.User
	for _, v := range organization.UserIDList {
		var user model.User
		objectID := bson.ObjectIdHex(v)
		err := organizationMongo.ConnectionDB.DB(dbName).C(collectionUser).FindId(objectID).One(&user)
		userList = append(userList, user)
	}
	return userList, err
}

// finish
// AddOrganization into database
func (organizationMongo *OrganizationRepositoryMongo) AddOrganization(organization model.Organization) (string, error) {
	organization.OrganizationID = bson.NewObjectId()
	err := organizationMongo.ConnectionDB.DB(dbName).C(collectionOrganization).Insert(organization)
	return organization.OrganizationID.Hex(), err
}

// Update new organization
func (organizationMongo *OrganizationRepositoryMongo) UpdateOrganization(organization model.Organization) (string, error) {
	err := organizationMongo.ConnectionDB.DB(dbName).C(collectionOrganization).RemoveId(bson.ObjectIdHex(organization.OrganizationID))
	return "string", err
}

//finish
func (organizationMongo *OrganizationRepositoryMongo) DeleteOrganization(orgId string) error {
	return organizationMongo.ConnectionDB.DB(dbName).C(collectionOrganization).RemoveId(bson.ObjectIdHex(orgId))
}
