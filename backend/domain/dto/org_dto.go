package dto

import (
	"backendSenior/domain/model"
	"common/utils/db"
)

// UpdateMeDto is request body for update user-info(model.USER)
type OrgDto struct {
	Name string `json:"name" validate:"required,gt=0"`
}

func (d *OrgDto) ToOrg() model.Organize {
	return model.Organize{
		Name:    d.Name,
		Members: nil,
		Admins:  nil,
		Rooms:   nil,
	}
}

type FindOrgByNameDto struct {
	Name string `json:"name" validate:"min=5"`
}

func (d *FindOrgByNameDto) ToFilter() model.OrganizationT {
	return model.OrganizationT{
		Name: db.Contains(d.Name, db.CaseInsensitive),
	}
}
