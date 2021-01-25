package repository

import "backendSenior/domain/model"

type ProxyMasterAPI interface {
	GetRoomMasterProxy(roomID string) (masterProxy model.Proxy, err error)
	GetProxyByID(proxyID string) (proxy model.Proxy, err error)
}
