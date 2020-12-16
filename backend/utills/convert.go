package utills

import "github.com/globalsign/mgo/bson"

// ToObjectIdArr convert objectID array to string array
func ToStringArr(objIdArr []bson.ObjectId) []string {
	var result = make([]string, len(objIdArr))
	n := len(objIdArr)
	for i := 0; i < n; i++ {
		result[i] = objIdArr[i].Hex()
	}
	return result
}

// ToObjectIdArr convert string array to objectID array
func ToObjectIdArr(stringArr []string) []bson.ObjectId {
	result := make([]bson.ObjectId, len(stringArr))
	n := len(stringArr)
	for i := 0; i < n; i++ {
		result[i] = bson.ObjectIdHex(stringArr[i])
	}
	return result
}
