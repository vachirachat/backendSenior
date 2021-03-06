package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/utills"
	"errors"
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/globalsign/mgo/txn"
)

type OrganizeUserRepositoryMongo struct {
	ConnectionDB *mgo.Session
	txnRunner    *txn.Runner
}

func NewOrganizeUserRepositoryMongo(conn *mgo.Session) *OrganizeUserRepositoryMongo {
	runner := txn.NewRunner(conn.DB(dbName).C(collectionTXNRoomUser))
	return &OrganizeUserRepositoryMongo{
		ConnectionDB: conn,
		txnRunner:    runner,
	}
}

var _ repository.OrganizeUserRepository = (*OrganizeUserRepositoryMongo)(nil)

func (repo *OrganizeUserRepositoryMongo) AddAdminToOrganize(orgID string, adminIDs []string) error {
	// Preconfition check
	n, err := repo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.M{"$in": utills.ToObjectIdArr(adminIDs)}).Count()

	if err != nil {
		return err
	}
	if n != len(adminIDs) {
		return fmt.Errorf("Invalid orgID, some of them not exists %d/%d", n, len(adminIDs))
	}

	n, err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(orgID)).Count()
	if n != 1 {
		return errors.New("Invalid orgID")
	}

	// Note: as per doc, every tranasction that update field that's managed by mgo/txn
	// must be update via mgo/txn.
	ops := []txn.Op{
		{
			C:  collectionOrganize,
			Id: bson.ObjectIdHex(orgID),
			Update: bson.M{
				"$addToSet": model.OrganizationT{
					Admins: bson.M{
						"$each": utills.ToObjectIdArr(adminIDs),
					},
					Members: bson.M{
						"$each": utills.ToObjectIdArr(adminIDs),
					},
				},
			},
		},
	}
	for _, adminID := range adminIDs {
		ops = append(ops, txn.Op{
			C:  collectionUser,
			Id: bson.ObjectIdHex(adminID),
			Update: bson.M{
				"$addToSet": model.UserUpdateMongo{
					Organize: bson.ObjectIdHex(orgID),
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

func (repo *OrganizeUserRepositoryMongo) AddMembersToOrganize(orgID string, memberIDs []string) error {
	// Preconfition check
	n, err := repo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.M{"$in": utills.ToObjectIdArr(memberIDs)}).Count()

	if err != nil {
		return err
	}
	if n != len(memberIDs) {
		return fmt.Errorf("Invalid organizeID, some of them not exists %d/%d", n, len(memberIDs))
	}

	n, err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(orgID)).Count()
	if n != 1 {
		return errors.New("Invalid organize ID")
	}

	ops := []txn.Op{
		{
			C:  collectionOrganize,
			Id: bson.ObjectIdHex(orgID),
			Update: bson.M{
				"$addToSet": model.OrganizationT{
					Members: bson.M{
						"$each": utills.ToObjectIdArr(memberIDs),
					},
				},
			},
		},
	}
	for _, memberID := range memberIDs {
		ops = append(ops, txn.Op{
			C:  collectionUser,
			Id: bson.ObjectIdHex(memberID),
			Update: bson.M{
				"$addToSet": model.UserUpdateMongo{
					Organize: bson.ObjectIdHex(orgID),
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

func (repo *OrganizeUserRepositoryMongo) DeleleOrganizeAdmin(orgID string, adminIDs []string) error {
	// Precondition check
	n, err := repo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.M{"$in": utills.ToObjectIdArr(adminIDs)}).Count()
	if err != nil {
		return err
	}
	if n != len(adminIDs) {
		return errors.New("Invalid adminIds, some of them not exists")
	}

	n, err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(orgID)).Count()
	if n != 1 {
		return errors.New("Invalid Organize ID")
	}

	// Note: as per doc, every tranasction that update field that's managed by mgo/txn
	// must be update via mgo/txn.
	ops := []txn.Op{
		{
			C:  collectionOrganize,
			Id: bson.ObjectIdHex(orgID),
			Update: bson.M{
				"$pullAll": model.OrganizationT{
					Admins: utills.ToObjectIdArr(adminIDs),
				},
			},
		},
	}

	err = repo.txnRunner.Run(ops, "", nil)
	if err != nil {
		return err
	}

	return nil
}

func (repo *OrganizeUserRepositoryMongo) DeleleOrganizeMember(orgID string, memberIDs []string) error {
	// Precondition check
	n, err := repo.ConnectionDB.DB(dbName).C(collectionUser).FindId(bson.M{"$in": utills.ToObjectIdArr(memberIDs)}).Count()
	if err != nil {
		return err
	}
	if n != len(memberIDs) {
		return errors.New("Invalid employeeIds, some of them not exists")
	}

	n, err = repo.ConnectionDB.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(orgID)).Count()
	if n != 1 {
		return errors.New("Invalid Organize ID")
	}

	ops := []txn.Op{
		{
			C:  collectionOrganize,
			Id: bson.ObjectIdHex(orgID),
			Update: bson.M{
				"$pullAll": model.OrganizationT{
					Members: utills.ToObjectIdArr(memberIDs),
					Admins:  utills.ToObjectIdArr(memberIDs),
				},
			},
		},
	}
	for _, memberID := range memberIDs {
		ops = append(ops, txn.Op{
			C:  collectionUser,
			Id: bson.ObjectIdHex(memberID),
			Update: bson.M{
				"$pull": model.UserUpdateMongo{
					Organize: bson.ObjectIdHex(orgID),
				},
			},
		})
	}

	err = repo.txnRunner.Run(ops, bson.NewObjectId(), nil)
	if err != nil {
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
