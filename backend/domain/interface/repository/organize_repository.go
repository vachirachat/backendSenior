package repository

import "backendSenior/domain/model"

// OrganizeRepository represent interface for managing Organize
type OrganizeRepository interface {
	GetAllOrganize() ([]model.Organize, error)
	CreateOrganize(organize model.Organize) (string, error)
	DeleteOrganize(organizeID string) error
	GetOrganizeById(organizeID string) (model.Organize, error)
	UpdateOrganize(organizeID string, name string) error
}

type OrganizeUserRepository interface {
	GetUserOrganizeById(userId string) ([]string, error)
	AddAdminToOrganize(organizeID string, adminIds []string) error
	DeleleOrganizeAdmin(organizeID string, adminIds []string) error
	AddMembersToOrganize(organizeID string, employeeIds []string) error
	DeleleOrganizeMember(organizeID string, adminIds []string) error
}