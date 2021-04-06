package mongo_repository

import "github.com/globalsign/mgo"

// KeyVersionCollection get collection for key version
func KeyVersionCollection(conn *mgo.Session) *mgo.Collection {
	return conn.DB(dbName).C(collectionKeyVersion)
}
