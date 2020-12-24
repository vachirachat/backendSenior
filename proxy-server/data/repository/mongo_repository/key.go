package mongo_repository

import (
	"crypto/rand"
	"errors"
	"fmt"
	"proxySenior/domain/interface/repository"
	model_proxy "proxySenior/domain/model"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type KeyRepository struct {
	ConnectionDB *mgo.Session
}

func NewKeyRepositoryMongo(conn *mgo.Session) *KeyRepository {
	return &KeyRepository{
		ConnectionDB: conn,
	}
}

// Temp.
var collectionRoomKey = "RoomKey"

var _ repository.Keystore = (*KeyRepository)(nil)

func (repo *KeyRepository) GetKeyForMessage(roomID string, timestamp time.Time) (model_proxy.RoomKeys, error) {
	//Now only key map
	var keyRoom model_proxy.RoomKeys
	cnt, err := repo.ConnectionDB.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).Count()
	if cnt != 0 || err == nil {
		err = repo.ConnectionDB.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).One(&keyRoom)
		if err != nil {
			return model_proxy.RoomKeys{}, errors.New("Error cannot get key : RoomID Please add key Frist")
		}
		return keyRoom, nil
	}
	return model_proxy.RoomKeys{}, errors.New("Error not found : Fail to Get")
}

func (repo *KeyRepository) GetKeyByRoom(roomID string) (keys []model_proxy.KeyRecord, err error) {
	var keyRoom model_proxy.RoomKeys
	cnt, err := repo.ConnectionDB.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).Count()
	if cnt == 0 || err != nil {
		err = repo.ConnectionDB.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).One(&keyRoom)
		if err != nil {
			return []model_proxy.KeyRecord{}, errors.New("Error cannot get key : RoomID Please add key Frist")
		}
		// Still hard code
		return keyRoom.KeyRecodes, nil
	}
	return []model_proxy.KeyRecord{}, errors.New("Error not found : RoomID Please add key Frist")
}

func (repo *KeyRepository) UpdateNewKey(roomID string, keyRecord []model_proxy.KeyRecord, timestamp time.Time) (model_proxy.RoomKeys, error) {
	var keyRoom model_proxy.RoomKeys
	cnt, err := repo.ConnectionDB.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).Count()
	if cnt == 0 || err != nil {
		if err != nil {
			fmt.Errorf("UpdateNewKey error: %s", err)
			return model_proxy.RoomKeys{}, err
		}
	}
	err = repo.ConnectionDB.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).One(&keyRoom)
	if err != nil {
		return model_proxy.RoomKeys{}, err
	}
	lastKey := keyRoom.KeyRecodes[len(keyRoom.KeyRecodes)-1]
	if !(lastKey.ValidTo.Before(timestamp) && lastKey.ValidFrom.After(timestamp)) {
		keyRoom, err = generateRoomKey(roomID, keyRecord)
		if err != nil {
			return model_proxy.RoomKeys{}, err
		}
		return keyRoom, repo.ConnectionDB.DB(dbName).C(collectionRoomKey).UpdateId(keyRoom.RoomID, keyRoom)
	}
	return model_proxy.RoomKeys{}, errors.New("Error not found : Fail to Update")

}

func (repo *KeyRepository) AddNewKey(roomID string, keyRecord []model_proxy.KeyRecord) (model_proxy.RoomKeys, error) {
	var keyRoom model_proxy.RoomKeys
	cnt, err := repo.ConnectionDB.DB(dbName).C(collectionRoomKey).FindId(bson.ObjectIdHex(roomID)).Count()
	if cnt == 0 || err != nil {
		//First time add
		keyRoom, err := generateRoomKey(roomID, keyRecord)
		if err != nil {
			return model_proxy.RoomKeys{}, err
		}
		return keyRoom, repo.ConnectionDB.DB(dbName).C(collectionRoomKey).Insert(keyRoom)

	}
	return keyRoom, errors.New("Room Already Exist")
}

// Secrect Just random string from now.
func generateRoomKey(roomID string, keyRecord []model_proxy.KeyRecord) (model_proxy.RoomKeys, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return model_proxy.RoomKeys{}, err
	}
	// Fix Must refractor
	keyRec := model_proxy.KeyRecord{
		Key:       key,
		ValidFrom: timeIn(BACKKOKTIMEZONE),
		ValidTo:   timeIn(BACKKOKTIMEZONE).Add(time.Minute * 5)}

	return model_proxy.RoomKeys{
		RoomID:     bson.ObjectIdHex(roomID),
		KeyRecodes: append(keyRecord, keyRec),
	}, nil
}

const BACKKOKTIMEZONE = "Asia/Bangkok"

func timeIn(name string) time.Time {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return time.Now().In(loc)
}
