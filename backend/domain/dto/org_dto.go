package dto

import (
	"backendSenior/domain/model"
	"common/utils/db"
)

type FindOrgByNameDto struct {
	Name string `json:"name" validate:"min=5"`
}

func (d *FindOrgByNameDto) ToFilter() model.OrganizationT {
	return model.OrganizationT{
		Name: db.Contains(d.Name, db.CaseInsensitive),
	}
}
