package db

import "github.com/globalsign/mgo/bson"

const (
	CaseInsensitive = "i"
)

func Contains(str string, flags string) bson.M {
	return bson.M{
		"$regex":   str,
		"$options": flags,
	}
}
