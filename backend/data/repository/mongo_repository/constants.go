package mongo_repository

import (
	"backendSenior/utills"

	"github.com/globalsign/mgo/bson"
)

const (
	dbName             = "mychat"
	collectionMessage  = "messages"
	collectionUser     = "users"
	collectionRoom     = "rooms"
	collectionProxy    = "proxies"
	collectionOrganize = "organize"
	collectionFCMToken = "fcmTokens"
	// for mgo/txn
	collectionTXNRoomUser = "txnRoomUser"
	collectionKeyVersion  = "keyVersions"
	// meta
	collectionMeta = "filemeta"
)

// return filter of {_id: {$in: ... }}, query that match multiple ID
func idInArr(ids []string) interface{} {
	return bson.M{
		"_id": inArr(ids),
	}
}

// return filter of {_id: {$in: ... }}, query that match multiple ID
func nameOrg(orgName string) interface{} {
	return bson.M{
		"name": orgName,
	}
}

// return filter of {$in: ... }, query that match multiple ID
func inArr(ids []string) interface{} {
	return bson.M{
		"$in": utills.ToObjectIdArr(ids),
	}
}
