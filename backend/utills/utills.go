package utills

import (
	"log"

	"github.com/globalsign/mgo/bson"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10) //salt 10
	if err != nil {
		log.Println("error HashPassword", err.Error())
		return ""
	}
	return string(bytes)
}

func RemoveFormListString(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func RemoveFormListBson(s []bson.ObjectId, r bson.ObjectId) []bson.ObjectId {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// ArrStringRemoveMatched return new slice with element in `arr` but not in `match`, and number of removed elements
func ArrStringRemoveMatched(arr []string, match []string) ([]string, int) {
	idx := 0
	n := len(arr)

	// set for quick look up
	set := make(map[string]bool)
	for _, v := range match {
		set[v] = true
	}

	result := make([]string, n)

	for i := 0; i < n; i++ {
		if _, exist := set[arr[i]]; !exist {
			result[idx] = arr[i]
			idx++
		}
	}
	// resize slice
	result = result[:idx]
	return result, n - idx
}

func ToObjectIdArr(stringArr []string) []bson.ObjectId {
	result := make([]bson.ObjectId, len(stringArr))
	n := len(stringArr)
	for i := 0; i < n; i++ {
		result[i] = bson.ObjectIdHex(stringArr[i])
	}
	return result
}
