package model

import (
	"github.com/globalsign/mgo/bson"
)

type UserInfo struct {
	User []User `json:"users"`
}

// type UserInfoSecrect struct {
// 	UserSecret []UserSecret `json:"users"`
// }

type UserTokenInfo struct {
	UserToken []UserToken `json:"users"`
}

type User struct {
	UserID    bson.ObjectId   `json:"userId" bson:"_id,omitempty"`
	Name      string          `json:"name" bson:"name,omitempty"`
	Email     string          `json:"email" bson:"email,omitempty"`
	Password  string          `json:"-" bson:"password,omitempty"`
	Room      []bson.ObjectId `json:"room" bson:"room,omitempty"`
	Organize  []bson.ObjectId `json:"organize" bson:"organize,omitempty"`
	UserType  string          `json:"userType" bson:"userType,omitempty"`
	FCMTokens []string        `json:"-" bson:"fcmTokens,omitempty"`
}

// UserInsert is used for inserting where empty fields are
// not omitted so that we can insert empty array to the database
type UserInsert struct {
	UserID    bson.ObjectId   `json:"userId" bson:"_id,omitempty"`
	Name      string          `json:"name" bson:"name"`
	Email     string          `json:"email" bson:"email"`
	Password  string          `json:"password" bson:"password"`
	Room      []bson.ObjectId `json:"room" bson:"room"`
	Organize  []bson.ObjectId `json:"organize" bson:"organize"`
	UserType  string          `json:"userType" bson:"userType"`
	FCMTokens []string        `json:"fcmTokens" bson:"fcmTokens"`
}

// // UserWithPassword is same as user but password isn't omitted in json
// type UserWithPassword struct {
// 	UserID    bson.ObjectId   `json:"userId" bson:"_id,omitempty"`
// 	Name      string          `json:"name" bson:"name,omitempty"`
// 	Email     string          `json:"email" bson:"email,omitempty"`
// 	Password  string          `json:"password" bson:"password,omitempty"`
// 	Room      []bson.ObjectId `json:"room" bson:"room,omitempty"`
// 	Organize  []bson.ObjectId `json:"organize" bson:"organize,omitempty"`
// 	UserType  string          `json:"userType" bson:"userType,omitempty"`
// 	FCMTokens []string        `json:"fcmTokens" bson:"fcmTokens,omitempty"`
// }

// UserUpdateMongo has same fields as user, but has types of interface{}.
// It's used instead of raw bson.M in update operations to ensure that when field name change in user model
// is always reflected
type UserUpdateMongo struct {
	UserID    interface{} `bson:"_id,omitempty"`
	Name      interface{} `bson:"name,omitempty"`
	Email     interface{} `bson:"email,omitempty"`
	Password  interface{} `bson:"password,omitempty"`
	Room      interface{} `bson:"room,omitempty"`
	Organize  interface{} `bson:"organize,omitempty"`
	UserType  interface{} `bson:"userType,omitempty"`
	FCMTokens interface{} `bson:"fcmTokens,omitempty"`
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
