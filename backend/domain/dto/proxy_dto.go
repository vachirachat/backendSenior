package dto

import (
	"backendSenior/domain/model"
)

// UpdateMeDto is request body for update user-info(model.USER)
type ProxyDto struct {
	Name string `json:"name" validate:"required,gt=0"`
}

func (d *ProxyDto) ToOrg() model.Organize {
	return model.Organize{
		Name:    d.Name,
		Members: nil,
		Admins:  nil,
		Rooms:   nil,
	}
}
