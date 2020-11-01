package model

type UserInfo struct {
	User []User `json:"users"`
}

type UserTokenInfo struct {
	UserToken []UserToken `json:"users"`
}

type User struct {
	Name     string   `json:"name" bson:"name"`
	Email    string   `json:"email" bson:"email"`
	Password string   `json:"password" bson:"password"`
	UserID   string   `json:"userID" bson:"userID"`
	Room     []string `json:"room" bson:"room"`
	UserType string   `json:"userType" bson:"userType"`
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
