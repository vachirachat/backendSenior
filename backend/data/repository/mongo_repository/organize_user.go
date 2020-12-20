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
)

type OrganizeUserRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

func NewOrganizeUserRepositoryMongo(conn *mgo.Session) *OrganizeUserRepositoryMongo {
	return &OrganizeUserRepositoryMongo{
		ConnectionDB: conn,
	}
}

var _ repository.OrganizeUserRepository = (*OrganizeUserRepositoryMongo)(nil)

func (repo *OrganizeUserRepositoryMongo) AddAdminToOrganize(organizeID string, adminIds []string) error {
	// Preconfition check
	n, err := repo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.M{"$in": utills.ToObjectIdArr(adminIds)}).Count()

	if err != nil {
		return err
	}
	if n != len(adminIds) {
		return fmt.Errorf("Invalid organizeID, some of them not exists %d/%d", n, len(adminIds))
	}

	n, err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(organizeID)).Count()
	if n != 1 {
		return errors.New("Invalid organizeID")
	}

	// Update database
	err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).UpdateId(bson.ObjectIdHex(organizeID), bson.M{
		"$addToSet": bson.M{
			"listAdmin": bson.M{
				"$each": utills.ToObjectIdArr(adminIds), // add all from listUser to array
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (repo *OrganizeUserRepositoryMongo) AddMembersToOrganize(organizeID string, employeeIds []string) error {
	// Preconfition check
	n, err := repo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.M{"$in": utills.ToObjectIdArr(employeeIds)}).Count()

	if err != nil {
		return err
	}
	if n != len(employeeIds) {
		return fmt.Errorf("Invalid organizeID, some of them not exists %d/%d", n, len(employeeIds))
	}

	n, err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(organizeID)).Count()
	if n != 1 {
		return errors.New("Invalid organize ID")
	}

	// Update database
	err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).UpdateId(bson.ObjectIdHex(organizeID), bson.M{
		"$addToSet": bson.M{
			"listMember": bson.M{
				"$each": utills.ToObjectIdArr(employeeIds), // add all from listUser to array
			},
		},
	})
	if err != nil {
		log.Println("AddMembersToOrganize : UpdateId Org. fail ", err)
		return err
	}

	_, err = repo.ConnectionDB.DB(dbName).C(collectionUser).UpdateAll(bson.M{"_id": bson.M{"$in": utills.ToObjectIdArr(employeeIds)}}, bson.M{
		"$addToSet": bson.M{
			"organize": bson.ObjectIdHex(organizeID),
		},
	})
	if err != nil {
		// TODO it should revert
		log.Println("AddMembersToOrganize : UpdateId Users fail ", err)
		return err
	}
	return nil
}

func (repo *OrganizeUserRepositoryMongo) DeleleOrganizeAdmin(organizeID string, adminIds []string) error {
	// Precondition check
	n, err := repo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.M{"$in": utills.ToObjectIdArr(adminIds)}).Count()
	if err != nil {
		return err
	}
	if n != len(adminIds) {
		return errors.New("Invalid adminIds, some of them not exists")
	}

	n, err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(organizeID)).Count()
	if n != 1 {
		return errors.New("Invalid Organize ID")
	}

	// Update database
	err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).UpdateId(bson.ObjectIdHex(organizeID), bson.M{
		"$pullAll": bson.M{
			"listAdmin": utills.ToObjectIdArr(adminIds),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (repo *OrganizeUserRepositoryMongo) DeleleOrganizeMember(organizeID string, employeeIds []string) error {
	// Precondition check
	n, err := repo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.M{"$in": utills.ToObjectIdArr(employeeIds)}).Count()
	if err != nil {
		return err
	}
	if n != len(employeeIds) {
		return errors.New("Invalid employeeIds, some of them not exists")
	}

	n, err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(organizeID)).Count()
	if n != 1 {
		return errors.New("Invalid Organize ID")
	}

	// Update database
	err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).UpdateId(bson.ObjectIdHex(organizeID), bson.M{
		"$pullAll": bson.M{
			"listMember": utills.ToObjectIdArr(employeeIds),
		},
	})
	if err != nil {
		return err
	}

	_, err = repo.ConnectionDB.DB(dbName).C(collectionUser).UpdateAll(bson.M{"_id": bson.M{"$in": utills.ToObjectIdArr(employeeIds)}}, bson.M{
		"$pull": bson.M{
			"organize": bson.ObjectIdHex(organizeID),
		},
	})
	if err != nil {
		// TODO it should revert
		return err
	}

	return nil
}

func (repo *OrganizeUserRepositoryMongo) GetUserOrganizeById(userId string) ([]string, error) {
	var user model.User
	err := repo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.ObjectIdHex(userId)).One(&user)
	if err != nil {
		return nil, err
	}
	return utills.ToStringArr(user.Organize), nil
}
