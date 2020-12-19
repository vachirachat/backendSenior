package repository

import "backendSenior/domain/model"

// ProxyRepository represent interface for managing proxy
type ProxyRepository interface {
	AddProxy(model.Proxy) (string, error)
	GetAllProxies() ([]model.Proxy, error)
	DeleteProxy(proxyID string) error
	UpdateProxy(proxyID string, update model.Proxy) error
	GetByID(proxyID string) (model.Proxy, error)
}
