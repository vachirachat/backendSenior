package mongo_repository

import "github.com/globalsign/mgo/bson"

func toObjectIdArr(stringArr []string) []bson.ObjectId {
	result := make([]bson.ObjectId, len(stringArr))
	n := len(stringArr)
	for i := 0; i < n; i++ {
		result[i] = bson.ObjectIdHex(stringArr[i])
	}
	return result
}
