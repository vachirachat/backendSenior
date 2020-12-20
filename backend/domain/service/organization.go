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

func (service *OrganizationService) GetAllOrganization() ([]model.Message, error) {
	messages, err := service.organizationRepo.GetAllOrganization(nil)
	return messages, err
}

func (service *OrganizationService) GetMemberInOrganization(orgID string) ([]model.user, error) {
	messages, err := service.organizationRepo.GetMemnerInOrganization(orgID)
	return messages, err
}

func (service *OrganizationService) AddOrganization(organization model.organization) (string, error) {
	msg, err := service.organizationRepo.AddOrganization(organization)
	return msg, err
}

func (service *OrganizationService) UpdateOrganization(organization model.organization) (string, error) {
	msgID, err := service.organizationRepo.UpdateOrganization(organization)
	return msgID, err
}

func (service *OrganizationService) DeleteOrganization(orgId string) error {
	err := service.organizationRepo.DeleteMessageByID(orgId)
	return err
}
