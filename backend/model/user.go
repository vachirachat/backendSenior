package model

import "github.com/globalsign/mgo/bson"

type UserInfo struct {
	User []User `json:"users"`
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
	Email       string `json:"email" bson:"email"`
	Token       string `json:"Token" bson:"Token"`
	TimeExpired string `json:"TimeExpired" bson:"TimeExpired"`
}

type UserLogin struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	isAdmin  bool   `json:"isadmin" bson:"isadmin"`
}
