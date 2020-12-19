package auth

import (
	"backendSenior/domain/interface/repository"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// ProxyAuth is the auth system for proxy
// it use username/password in header as authentication method
type ProxyAuth struct {
	proxyRepo repository.ProxyRepository
}

func NewProxyAuth(proxyRepo *repository.ProxyRepository) *ProxyAuth {
	return &ProxyAuth{
		proxyRepo: *proxyRepo,
	}
}

// VerifyCredentials verify clientSecret of proxy
func (auth *ProxyAuth) VerifyCredentials(clientID string, clientSecret string) (bool, error) {
	proxy, err := auth.proxyRepo.GetByID(clientID)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(proxy.Secret), []byte(clientSecret))
	if err != nil {
		return false, fmt.Errorf("invalid hash: %s", err.Error())
	}
	return true, nil
}
