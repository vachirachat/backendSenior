package repository

import (
	"backendSenior/domain/model"
)

type OrganizationRepository interface {
	GetAllOrganization() ([]model.Organization, error)
	GetMemberInOrganization(orgID string) ([]model.User, error)
	AddOrganization(organization model.Organization) (string, error)
	UpdateOrganization(organization model.Organization) (string, error)
	DeleteOrganization(orgID string) error
}