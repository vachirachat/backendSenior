package utills

import (
	"log"

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
