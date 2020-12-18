package model

import (
	"github.com/globalsign/mgo/bson"
)

type UserInfo struct {
	User []User `json:"users"`
}

type UserInfoSecrect struct {
	UserSecret []UserSecret `json:"users"`
}

type UserTokenInfo struct {
	UserToken []UserToken `json:"users"`
}

type User struct {
	UserID   bson.ObjectId   `json:"userID" bson:"_id,omitempty"`
	Name     string          `json:"name" bson:"name"`
	Email    string          `json:"email" bson:"email"`
	Password string          `json:"password" bson:"password"`
	Room     []bson.ObjectId `json:"room" bson:"room"`
	UserType string          `json:"userType" bson:"userType"`
}

type UserToken struct {
	UserID      bson.ObjectId `json:"userID" bson:"_id,omitempty"`
	Token       string        `json:"token" bson:"token"`
	TimeExpired string        `json:"timeexpired" bson:"timeexpired"`
}

type UserSecret struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
	Role     string `json:"role" bson:"role"`
}
