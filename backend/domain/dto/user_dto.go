package dto

import (
	"backendSenior/domain/model"
	"backendSenior/utills"

	"github.com/globalsign/mgo/bson"
)

// UpdateMeDto is request body for update user-info(model.USER)
type UpdateMeDto struct {
	Name     string `json:"name" validate:"required,gt=0"`
	Email    string `json:"email" validate:"required,gt=1,email"`
	UserType string `json:"userType" validate:"required,gt=1,eq=user"`
}

func (d *UpdateMeDto) ToUser() model.User {
	return model.User{
		Name:      d.Name,
		Email:     d.Email,
		Room:      nil,
		Organize:  nil,
		UserType:  d.UserType,
		FCMTokens: nil,
	}
}

type CreateUser struct {
	Name     string `json:"name" validate:"required,gt=0"`
	Email    string `json:"email" validate:"required,gt=1,email"`
	Password string `json:"password" validate:"required,gt=7"`
}

// init All user must be user-role for now
func (d *CreateUser) ToUser(isDashboard bool) model.User {
	role := "user"
	if isDashboard {
		role = "admin"
	}
	return model.User{
		Name:      d.Name,
		Email:     d.Email,
		Password:  utills.HashPassword(d.Password),
		Room:      []bson.ObjectId{},
		Organize:  []bson.ObjectId{},
		UserType:  role, // Fix set as user for test
		FCMTokens: []string{},
	}
}

type CreateUserSecret struct {
	Email    string `json:"email" validate:"required,gt=1,email"`
	Password string `json:"password" validate:"required,gt=7"`
}
