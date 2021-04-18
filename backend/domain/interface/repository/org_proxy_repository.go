package repository

// OrgRoomRepository is repository for managing org-room relation
type OrgProxyRepository interface {
	GetOrgProxyIDs(orgID string) (proxiseIDs []string, err error)
	AddProxiseToOrg(orgID string, proxyIDs []string) (err error)
	RemoveProxiseFromOrg(orgID string, proxyIDs []string) (err error)
}
