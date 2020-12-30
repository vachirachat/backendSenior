package mongo_repository

import (
	"backendSenior/domain/model"
	"backendSenior/utills"
	"errors"
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/globalsign/mgo/txn"
)

type OrgRoomRepository struct {
	conn      *mgo.Session
	txnRunner *txn.Runner
}

func NewOrgRoomRepository(conn *mgo.Session) *OrgRoomRepository {
	runner := txn.NewRunner(conn.DB(dbName).C(collectionTXNRoomUser))
	return &OrgRoomRepository{
		conn:      conn,
		txnRunner: runner,
	}
}

// GetOrgRooms return roomIDs of org
func (repo *OrgRoomRepository) GetOrgRooms(orgID string) ([]string, error) {
	var org model.Organize
	err := repo.conn.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(orgID)).One(&org)
	if err != nil {
		return nil, err
	}
	return utills.ToStringArr(org.Rooms), nil
}

// AddRoomsToOrg adds rooms to the org, each room must don't have any org
func (repo *OrgRoomRepository) AddRoomsToOrg(orgID string, roomIDs []string) error {
	n, err := repo.conn.DB(dbName).C(collectionRoom).Find(idInArr(roomIDs)).Count()

	if err != nil {
		return err
	}
	if n != len(roomIDs) {
		return fmt.Errorf("Invalid userIDs, some of them not exists %d/%d", n, len(roomIDs))
	}

	n, err = repo.conn.DB(dbName).C(collectionOrganize).FindId(bson.ObjectIdHex(orgID)).Count()
	if err != nil || n != 1 {
		return errors.New("Invalid Room ID")
	}

	ops := []txn.Op{
		{
			C:  collectionOrganize,
			Id: bson.ObjectIdHex(orgID),
			Update: bson.M{
				"$addToSet": model.OrganizationUpdateMongo{
					Rooms: bson.M{
						"$each": utills.ToObjectIdArr(roomIDs),
					},
				},
			},
		},
	}
	for _, roomID := range roomIDs {
		ops = append(ops, txn.Op{
			C:  collectionRoom,
			Id: bson.ObjectIdHex(roomID),
			Update: bson.M{
				"$set": model.RoomUpdateMongo{
					OrgID: bson.ObjectIdHex(orgID),
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
