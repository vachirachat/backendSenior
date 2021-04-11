package db

import (
	"github.com/globalsign/mgo/bson"
	"regexp"
)

const (
	CaseInsensitive = "i"
)

var escapePattern = regexp.MustCompile(`([-[\]{}()*+?.,/^$|#])`)

func escapeRe(re string) string {
	return escapePattern.ReplaceAllString(re, "\\$1")
}

func Contains(str string, flags string) bson.M {
	return bson.M{
		"$regex":   escapeRe(str),
		"$options": flags,
	}
}

func Set(update interface{}) bson.M {
	return bson.M{
		"$set": update,
	}
}
