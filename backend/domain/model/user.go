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
	UserID    bson.ObjectId   `json:"userId,omitempty" bson:"_id,omitempty"`
	Name      string          `json:"name,omitempty" bson:"name,omitempty"`
	Email     string          `json:"email,omitempty" bson:"email,omitempty"`
	Password  string          `json:"-" bson:"password,omitempty"`
	Room      []bson.ObjectId `json:"room,omitempty" bson:"room,omitempty"`
	Organize  []bson.ObjectId `json:"organize,omitempty" bson:"organize,omitempty"`
	UserType  string          `json:"userType,omitempty" bson:"userType,omitempty"`
	FCMTokens []string        `json:"fcmTokens,omitempty" bson:"fcmTokens,omitempty"`
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
