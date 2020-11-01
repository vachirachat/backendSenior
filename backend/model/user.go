package model

type UserInfo struct {
	User []User `json:"users"`
}

type User struct {
	Name     string   `json:"name" bson:"name"`
	Email    string   `json:"email" bson:"email"`
	Password string   `json:"password" bson:"password"`
	UserID   string   `json:"userID" bson:"userID"`
	Room     []string `json:"room" bson:"room"`
	UserType string   `json:"userType" bson:"userType"`
}
