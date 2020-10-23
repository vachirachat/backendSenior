package model

type User struct {
	Name         string    `json:"name" bson:"name"`
	Room         []string  `json:"room" bson:"room"`
	RoomAdmit    []string  `json:"roomAdmit" bson:"roomAdmit"`
	UserID       string    `json:"userID" bson:"userID"`
	Notification []Message `json:"notification" bson:"notification`
}
