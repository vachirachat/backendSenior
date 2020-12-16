package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
)

// ProxyService provide acces sto proxy related function
type ProxyService struct {
	proxyRepo repository.ProxyRepository
}

// NewProxyService create nenw instance of `ProxyService`
func NewProxyService(proxyRepo repository.ProxyRepository) *ProxyService {
	return &ProxyService{
		proxyRepo: proxyRepo,
	}
}

// NewProxy create new proxy with name (display name)
func (service *ProxyService) NewProxy(name string) (string, error) {
	return service.proxyRepo.AddProxy(name)
}

// GetAll return list of all proxies
func (service *ProxyService) GetAll() ([]model.Proxy, error) {
	return service.proxyRepo.GetAllProxies()
}

// EditProxyName change proxy name
func (service *ProxyService) EditProxyName(proxyID string, name string) error {
	return service.proxyRepo.UpdateProxy(proxyID, model.Proxy{
		Name: name,
	})
}

// GetProxyByID return proxy with specified ID
func (service *ProxyService) GetProxyByID(proxyID string) (model.Proxy, error) {
	return service.proxyRepo.GetByID(proxyID)
}

// DeleteProxyByID delte proxy with specified ID
func (service *ProxyService) DeleteProxyByID(proxyID string) error {
	return service.proxyRepo.DeleteProxy(proxyID)
}
