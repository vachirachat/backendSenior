package model

import "github.com/globalsign/mgo/bson"

type UserInfo struct {
	User []User `json:"users"`
}

type UserInfoSecrect struct {
	UserLogin []UserLogin `json:"users"`
}

type UserTokenInfo struct {
	UserToken []UserToken `json:"users"`
}

// type User struct {
// 	UserID   bson.ObjectId   `json:"userID" bson:"_id,omitempty"`
// 	Name     string          `json:"name" bson:"name"`
// 	Email    string          `json:"email" bson:"email"`
// 	Password string          `json:"password" bson:"password"`
// 	Room     []bson.ObjectId `json:"room" bson:"room"`
// 	UserType string          `json:"userType" bson:"userType"`
// }

// // User is same as user expect all objectID are now string
type User struct {
	UserID   string   `json:"userID" bson:"_id,omitempty"`
	Name     string   `json:"name" bson:"name"`
	Email    string   `json:"email" bson:"email"`
	Password string   `json:"password" bson:"password"`
	Room     []string `json:"room" bson:"room"`
	UserType string   `json:"userType" bson:"userType"`
}

type UserToken struct {
	Email       string `json:"email" bson:"email"`
	Token       string `json:"Token" bson:"Token"`
	TimeExpired string `json:"TimeExpired" bson:"TimeExpired"`
}

type UserLogin struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
	isAdmin  bool   `json:"isadmin" bson:"isadmin"`
}

// Re-Assign byte string(From mondo bson.ObjectID) to String
func (user *User) UserStringIDToMongoID() User {
	user.UserID = bson.ObjectId(user.UserID).Hex()
	user.Room = ArrUserListObjectToString(user.Room)
	return *user
}

func ArrUserListObjectToString(rooms []string) []string {
	for i := range rooms {
		rooms[i] = bson.ObjectId(rooms[i]).Hex()
	}
	return rooms
}

func ArrUserMongoToRoomString(users []User) []User {
	for i := range users {
		users[i].UserID = bson.ObjectId(users[i].UserID).Hex()
		users[i].Room = ArrUserListObjectToString(users[i].Room)
	}
	return users
}
