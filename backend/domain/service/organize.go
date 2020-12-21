package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
)

// OrganizeService provide acces sto proxy related function
type OrganizeService struct {
	organizeRepo     repository.OrganizeRepository
	organizeUserRepo repository.OrganizeUserRepository
}

// NewOrganizeService create nenw instance of `OrganizeService`
func NewOrganizeService(organizeRepo repository.OrganizeRepository, organizeUserRepo repository.OrganizeUserRepository) *OrganizeService {
	return &OrganizeService{
		organizeRepo:     organizeRepo,
		organizeUserRepo: organizeUserRepo,
	}
}

// NewOrganize create new Organize with name (display name)
func (service *OrganizeService) AddOrganize(organize model.Organize) (string, error) {
	return service.organizeRepo.CreateOrganize(organize)
}

// GetAll return list of all Organize
func (service *OrganizeService) GetAllOrganizes() ([]model.Organize, error) {
	return service.organizeRepo.GetAllOrganize()
}

// GetOrganizeByID return Organize with specified ID
func (service *OrganizeService) GetOrganizeById(organizeID string) (model.Organize, error) {
	return service.organizeRepo.GetOrganizeById(organizeID)
}

// DeleteOrganize delete Organize with specified ID
func (service *OrganizeService) DeleteOrganizeByID(organizeID string) error {
	return service.organizeRepo.DeleteOrganize(organizeID)
}

// EditOrganizeName change Organize name
func (service *OrganizeService) EditOrganizeName(proxyID string, organize model.Organize) error {
	return service.organizeRepo.UpdateOrganize(proxyID, organize.Name)
}

// Add Organize adminIds
func (service *OrganizeService) AddAdminToOrganize(name string, adminIds []string) error {
	return service.organizeUserRepo.AddAdminToOrganize(name, adminIds)
}

// Remove adminIds Organize
func (service *OrganizeService) DeleteAdminFromOrganize(organizeID string, adminIds []string) error {
	return service.organizeUserRepo.DeleleOrganizeAdmin(organizeID, adminIds)
}

// Invite employeeIds Organize
func (service *OrganizeService) AddMemberToOrganize(organizeID string, employeeIds []string) error {
	return service.organizeUserRepo.AddMembersToOrganize(organizeID, employeeIds)
}

// Remove employeeIds Organize
func (service *OrganizeService) DeleteMemberFromOrganize(organizeID string, employeeIds []string) error {
	return service.organizeUserRepo.DeleleOrganizeMember(organizeID, employeeIds)
}

// Remove employeeIds Organize
func (service *OrganizeService) GetMyOrganize(userId string) ([]string, error) {
	return service.organizeUserRepo.GetUserOrganizeById(userId)
}
