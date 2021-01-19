package mongo_repository

import (
	"encoding/json"
	"fmt"
	"proxySenior/domain/interface/repository"
	model_proxy "proxySenior/domain/model"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type KeyRepository struct {
	conn *mgo.Session
	col  *mgo.Collection
}

// Temp.
var collectionRoomKey = "keys"

func NewKeyRepositoryMongo(conn *mgo.Session) *KeyRepository {
	return &KeyRepository{
		conn: conn,
		col:  conn.DB(dbName).C(collectionRoomKey),
	}
}

var _ repository.Keystore = (*KeyRepository)(nil)

// Find find keys according to the filter
func (r *KeyRepository) Find(filter interface{}) ([]model_proxy.KeyRecord, error) {
	var keys []model_proxy.KeyRecord
	err := r.col.Find(filter).All(&keys)

	return keys, err
}

// FindByRoom is shortcut for finding by room, it also sort by time descending
func (r *KeyRepository) FindByRoom(roomID string) ([]model_proxy.KeyRecord, error) {
	var keys []model_proxy.KeyRecord
	err := r.col.Find(model_proxy.KeyRecordUpdate{
		RoomID: bson.ObjectIdHex(roomID),
	}).Sort("-from").All(&keys)

	return keys, err
}

// AddNewKey add new keys to the room, also invalidate last key if available
func (r *KeyRepository) AddNewKey(roomID string, key []byte) error {
	now := time.Now()
	var dummy json.RawMessage

	var keys []model_proxy.KeyRecord
	_ = r.col.Find(model_proxy.KeyRecordUpdate{
		RoomID: bson.ObjectIdHex(roomID),
	}).All(&keys)

	// TODO make it transaction
	_, err := r.col.Find(model_proxy.KeyRecordUpdate{
		RoomID: bson.ObjectIdHex(roomID),
	}).Sort("-from").Limit(1).Apply(mgo.Change{
		Update: bson.M{
			"$set": model_proxy.KeyRecordUpdate{
				ValidTo: now,
			},
		},
		Upsert: false,
	}, &dummy)

	if err != nil && err.Error() != "not found" {
		return err
	}

	keyRecord := model_proxy.KeyRecord{
		Key:       key,
		RoomID:    bson.ObjectIdHex(roomID),
		ValidFrom: now,
	}
	err = r.col.Insert(keyRecord)

	return err
}

// ReplaceKey for replacing key in the room, when update
func (repo *KeyRepository) ReplaceKey(roomID string, keys []model_proxy.KeyRecord) error {
	// delete key
	// TODO: preserve old one  and revert if fail
	_, err := repo.col.RemoveAll(model_proxy.KeyRecordUpdate{RoomID: bson.ObjectIdHex(roomID)})
	if err != nil {
		return fmt.Errorf("error removing key: %w", err)
	}

	all := make([]interface{}, 0)
	for _, k := range keys {
		all = append(all, k)
	}

	err = repo.col.Insert(all...)
	if err != nil {
		return fmt.Errorf("error inserting key: %w", err)
	}

	return nil
}

// func (repo *KeyRepository) GetKeyByRoom(roomID string) (keys []model_proxy.KeyRecord, err error) {
// 	var keyRoom model_proxy.RoomKeys
// 	cnt, err := repo.conn.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).Count()
// 	if cnt == 0 || err != nil {
// 		err = repo.conn.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).One(&keyRoom)
// 		if err != nil {
// 			return []model_proxy.KeyRecord{}, errors.New("Error cannot get key : RoomID Please add key Frist")
// 		}
// 		// Still hard code
// 		return keyRoom.KeyRecords, nil
// 	}
// 	return []model_proxy.KeyRecord{}, errors.New("Error not found : RoomID Please add key Frist")
// }

// func (repo *KeyRepository) UpdateNewKey(roomID string, keyRecord []model_proxy.KeyRecord, timestamp time.Time) (model_proxy.RoomKeys, error) {
// 	var keyRoom model_proxy.RoomKeys
// 	cnt, err := repo.conn.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).Count()
// 	if cnt == 0 || err != nil {
// 		if err != nil {
// 			fmt.Errorf("UpdateNewKey error: %s", err)
// 			return model_proxy.RoomKeys{}, err
// 		}
// 	}
// 	err = repo.conn.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).One(&keyRoom)
// 	if err != nil {
// 		return model_proxy.RoomKeys{}, err
// 	}
// 	lastKey := keyRoom.KeyRecords[len(keyRoom.KeyRecords)-1]
// 	if !(lastKey.ValidTo.Before(timestamp) && lastKey.ValidFrom.After(timestamp)) {
// 		keyRoom, err = generateRoomKey(roomID, keyRecord)
// 		if err != nil {
// 			return model_proxy.RoomKeys{}, err
// 		}
// 		return keyRoom, repo.conn.DB(dbName).C(collectionRoomKey).UpdateId(keyRoom.RoomID, keyRoom)
// 	}
// 	return model_proxy.RoomKeys{}, errors.New("Error not found : Fail to Update")

// }

// func (repo *KeyRepository) AddNewKey(roomID string, keyRecord []model_proxy.KeyRecord) (model_proxy.RoomKeys, error) {
// 	var keyRoom model_proxy.RoomKeys
// 	cnt, err := repo.conn.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).Count()
// 	if cnt == 0 || err != nil {
// 		//First time add
// 		keyRoom, err := generateRoomKey(roomID, keyRecord)
// 		if err != nil {
// 			return model_proxy.RoomKeys{}, err
// 		}
// 		return keyRoom, repo.conn.DB(dbName).C(collectionRoomKey).Insert(keyRoom)

// 	}
// 	return keyRoom, errors.New("Room Already Exist")
// }
