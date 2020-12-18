package model

import (
	"github.com/globalsign/mgo/bson"
)

func ToStringArr(arrObject []bson.ObjectId) []string {
	var arrString = make([]string, len(arrObject))
	for i := range arrObject {
		arrString[i] = arrObject[i].Hex()
	}
	return arrString
}
