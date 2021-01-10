package repository

import "backendSenior/domain/model"

type ProxyMasterAPI interface {
	GetRoomMasterProxy(roomID string) (masterProxy model.Proxy, err error)
	GetProxyMasterRooms(proxyID string) (roomID []string, err error)
}
