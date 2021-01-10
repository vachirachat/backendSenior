package repository

import "backendSenior/domain/model"

type ProxyMasterRepo interface {
	GetRoomMasterProxy(roomID string) (masterProxy model.Proxy, err error)
	SetRoomMasterProxy(roomID string, masterProxyID string) (err error)
	GetProxyMasterRooms(proxyID string) (roomID []string, err error)
}
