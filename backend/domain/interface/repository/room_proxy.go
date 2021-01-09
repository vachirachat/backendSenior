package repository

import "backendSenior/domain/model"

// RoomProxyRepository is interface for repository managing room/proxy relation
type RoomProxyRepository interface {
	GetProxyRooms(proxyID string) (roomIDs []string, err error)
	GetRoomProxies(roomID string) (proxyIDs []string, err error)
	AddProxiesToRoom(roomID string, proxyIDs []string) (err error)
	RemoveProxiesFromRoom(roomID string, proxyIDs []string) (err error)

	GetRoomMasterProxy(roomID string) (masterProxy model.Proxy, err error)
	SetRoomMasterProxy(roomID string, masterProxyID string) (err error)
	GetProxyMasterRooms(proxyID string) (roomID []string, err error)
}
