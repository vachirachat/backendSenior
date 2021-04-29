package repository

import "backendSenior/domain/model"

// OrgRoomRepository is repository for managing org-room relation
type OrgProxyRepository interface {
	GetOrgProxyIDs(orgID string) (proxiseIDs []model.Proxy, err error)
	AddProxiseToOrg(orgID string, proxyIDs []string) (err error)
	RemoveProxiseFromOrg(orgID string, proxyIDs []string) (err error)
}
