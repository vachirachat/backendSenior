package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
)

// OrganizeService provide acces sto proxy related function
type OrganizeService struct {
	organizeRepo     repository.OrganizeRepository
	organizeUserRepo repository.OrganizeUserRepository
	orgRoomRepo      repository.OrgRoomRepository
}

// NewOrganizeService create nenw instance of `OrganizeService`
func NewOrganizeService(organizeRepo repository.OrganizeRepository, organizeUserRepo repository.OrganizeUserRepository, orgRoomRepo repository.OrgRoomRepository) *OrganizeService {
	return &OrganizeService{
		organizeRepo:     organizeRepo,
		organizeUserRepo: organizeUserRepo,
		orgRoomRepo:      orgRoomRepo,
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

// GetOrganizeByID return Organize with specified ID
func (service *OrganizeService) GetOrganizationsByIDs(organizeIDs []string) ([]model.Organize, error) {
	return service.organizeRepo.GetOrganizesByIDs(organizeIDs)
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

// GetUserOrganizeIDs return organization IDs of a user
func (service *OrganizeService) GetUserOrganizeIDs(userId string) ([]string, error) {
	return service.organizeUserRepo.GetUserOrganizeById(userId)
}

// GetUserOrganizations return all organization (object) of a user, it's shortcut for getting orgIDs from userID, then query by IDs
func (service *OrganizeService) GetUserOrganizations(userId string) ([]model.Organize, error) {
	orgIDs, err := service.GetUserOrganizeIDs(userId)
	if err != nil {
		return nil, err
	}
	orgs, err := service.organizeRepo.GetOrganizesByIDs(orgIDs)
	return orgs, err
}

// AddRoomsToOrg add rooms to the organizations, fail when either of rooms have already an org
func (service *OrganizeService) AddRoomsToOrg(orgID string, roomIDs []string) error {
	err := service.orgRoomRepo.AddRoomsToOrg(orgID, roomIDs)
	return err
}

// DeleteRoomsFromOrg remove rooms from org
func (service *OrganizeService) DeleteRoomsFromOrg(orgID string, roomIDs []string) error {
	err := service.orgRoomRepo.RemoveRoomsFromOrg(orgID, roomIDs)
	return err
}

// GetOrgRoomIDs return roomIDs of org
func (service *OrganizeService) GetOrgRoomIDs(orgID string) ([]string, error) {
	roomIDs, err := service.orgRoomRepo.GetOrgRooms(orgID)
	return roomIDs, err
}
