package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/globalsign/mgo/bson"
	"golang.org/x/crypto/bcrypt"
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
// return ID, secret, error
func (service *ProxyService) NewProxy(name string) (string, string, error) {
	randByte := make([]byte, 48)
	_, err := io.ReadFull(rand.Reader, randByte)
	if err != nil {
		return "", "", fmt.Errorf("generating secret: %s", err.Error())
	}
	secret := base64.StdEncoding.EncodeToString(randByte)
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("hashing secret: %s", err.Error())
	}

	id, err := service.proxyRepo.AddProxy(model.Proxy{
		Name:   name,
		Secret: string(hashedSecret),
		Rooms:  []bson.ObjectId{},
	})
	if err != nil {
		return "", "", fmt.Errorf("inserting proxy: %s", err.Error())
	}
	return id, secret, nil
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

// ResetProxySecret generate new secret for proxy
func (service *ProxyService) ResetProxySecret(proxyID string) (string, error) {
	randByte := make([]byte, 48)
	_, err := io.ReadFull(rand.Reader, randByte)
	if err != nil {
		return "", fmt.Errorf("generating secret: %s", err.Error())
	}
	secret := base64.StdEncoding.EncodeToString(randByte)
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing secret: %s", err.Error())
	}
	err = service.proxyRepo.UpdateProxy(proxyID, model.Proxy{
		Secret: string(hashedSecret),
	})
	if err != nil {
		return "", fmt.Errorf("updating proxy: %s", err.Error())
	}
	return secret, nil
}

// GetProxyByID return proxy with specified ID
func (service *ProxyService) GetProxyByID(proxyID string) (model.Proxy, error) {
	return service.proxyRepo.GetByID(proxyID)
}

// GetProxiesByIDs return proxy with specified ID
func (service *ProxyService) GetProxiesByIDs(proxyIDs []string) ([]model.Proxy, error) {
	return service.proxyRepo.GetByIDs(proxyIDs)
}

// DeleteProxyByID delte proxy with specified ID
func (service *ProxyService) DeleteProxyByID(proxyID string) error {
	return service.proxyRepo.DeleteProxy(proxyID)
}
