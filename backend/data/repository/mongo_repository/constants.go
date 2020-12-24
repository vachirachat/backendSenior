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
)

// return filter of {_id: {$in: ... }}, query that match multiple ID
func idInArr(ids []string) interface{} {
	return bson.M{
		"_id": bson.M{
			"$in": utills.ToObjectIdArr(ids),
		},
	}
}
