package repository

// OrgRoomRepository is repository for managing org-room relation
type OrgRoomRepository interface {
	GetOrgRooms(orgID string) (orgIDs []string, err error)
	AddRoomsToOrg(orgID string, roomIDs []string) (err error)
}
