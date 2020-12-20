package repository

import (
	"backendSenior/domain/model"
)

type OrganizationRepository interface {
	GetAllOrganization() ([]model.organization, error)
	GetMemberInOrganization(orgID string) ([]model.user, error)
	AddOrganization(organization model.organization) (string, error)
	UpdateOrganization(organization model.organization) (string, error)
	DeleteOrganization(orgID string) error
}