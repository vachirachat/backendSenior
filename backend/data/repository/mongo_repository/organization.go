package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"fmt"
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
func (messageMongo *OrganizationRepositoryMongo) GetAllOrganization() ([]model.Message, error) {
	var organizations []model.Organization
	err := organizationMongo.ConnectionDB.DB(dbName).C(collectionOrganization).Find(nil).All(&organizations)
	return organizations, err
}
//finish
// GetMessagesByRoom return messages from specified room, with optional time filter
func (messageMongo *OrganizationRepositoryMongo) GetMemberInOrganization(orgID string) ([]model.user, error) {
	var organization model.Organization
	err := organizationMongo.ConnectionDB.DB(dbName).C(collectionOrganization).Find(bson.ObjectIdHex(orgID)).All(&organization)
	log.Println(organization)
	return organization.UserIDList, err
}
// finish
// AddOrganization into database
func (messageMongo *OrganizationRepositoryMongo) AddOrganization(organization model.organization) (string, error) {
	organization.OrganizationID = bson.NewObjectId()
	err = organizationMongo.ConnectionDB.DB(dbName).C(collectionOrganization).Insert(organization)
	return organization.OrganizationID.Hex(), err
}

// Update new organization
func (messageMongo *OrganizationRepositoryMongo) UpdateOrganization(organization model.organization) (string, error) {
	return organizationMongo.ConnectionDB.DB(dbName).C(collectionOrganization).RemoveId(bson.ObjectIdHex(messageID))
}

//finish
func (messageMongo *OrganizationRepositoryMongo) DeleteOrganization(orgId string) error {
	return organizationMongo.ConnectionDB.DB(dbName).C(collectionOrganization).RemoveId(bson.ObjectIdHex(orgID))
}
