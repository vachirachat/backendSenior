package dto

import (
	"backendSenior/domain/model"
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
func (d *CreateUser) ToUser() model.User {
	return model.User{
		Name:      d.Name,
		Email:     d.Email,
		Password:  d.Password,
		Room:      nil,
		Organize:  nil,
		UserType:  "user",
		FCMTokens: nil,
	}
}

type CreateUserSecret struct {
	Email    string `json:"email" validate:"required,gt=1,email"`
	Password string `json:"password" validate:"required,gt=7"`
}
