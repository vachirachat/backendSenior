package model

type UserInfo struct {
	User []User `json:"users"`
}

type UserTokenInfo struct {
	UserToken []UserToken `json:"users"`
}

type User struct {
	UserID       string    `json:"userID" bson:"_id"`
	Name         string    `json:"name" bson:"name"`
	Room         []string  `json:"room" bson:"room"`
	RoomAdmit    []string  `json:"roomAdmit" bson:"roomAdmit"`
	Notification []Message `json:"notification" bson:"notification`
}

type UserToken struct {
	UserID      string `json:"userID" bson:"_id"`
	Token       string `json:"Token" bson:"Token"`
	TimeExpired string `json:"TimeExpired" bson:"TimeExpired"`
}

type UserLogin struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}
