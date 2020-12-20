package service

import (
	"backendSenior/domain/interface/repository"

	"backendSenior/domain/model"
)

// OrganizationService message service provide access to message related functions
type OrganizationService struct {
	organizationRepo repository.OrganizationRepository
}

// NewOrganizationService create message service from repository
func NewOrganizationService(orgRepo repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{
		organizationRepo: orgRepo,
	}
}

func (service *OrganizationService) GetAllOrganization() ([]model.Organization, error) {
	organizations, err := service.organizationRepo.GetAllOrganization()
	return organizations, err
}

func (service *OrganizationService) GetMemberInOrganization(orgID string) ([]model.User, error) {
	users, err := service.organizationRepo.GetMemberInOrganization(orgID)
	return users, err
}

func (service *OrganizationService) AddOrganization(organization model.Organization) (string, error) {
	org, err := service.organizationRepo.AddOrganization(organization)
	return org, err
}

func (service *OrganizationService) UpdateOrganization(organization model.Organization) (string, error) {
	msgID, err := service.organizationRepo.UpdateOrganization(organization)
	return msgID, err
}

func (service *OrganizationService) DeleteOrganization(orgId string) error {
	err := service.organizationRepo.DeleteOrganization(orgId)
	return err
}
