package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"errors"
	"time"

	"github.com/globalsign/mgo/bson"
)

type NotificationService struct {
	fcmRepo     repository.FCMTokenRepository
	fmcUserRepo repository.FCMUserRepository
}

var err_not_found = errors.New("not found")

// RegisterDevice register user device (specified by token) for receiving notificiation for that user
func (service *NotificationService) RegisterDevice(userID string, newToken string) error {
	_, err := service.fcmRepo.GetTokenByID(newToken)
	// found token
	if err == nil {
		return errors.New("token already registered")
	}

	// some other error
	if err != nil && err != err_not_found {
		return err
	}

	err = service.fcmRepo.AddToken(model.FCMToken{
		Token:       newToken,
		UserID:      bson.ObjectIdHex(userID),
		LastUpdated: time.Now(),
	})

	return err
}

// RefreshDevice refresh last update timestamp of a token, to prevent it from expiring
func (service *NotificationService) RefreshDevice(token string) error {
	err := service.fcmRepo.UpdateToken(token, model.FCMToken{
		LastUpdated: time.Now(),
	})

	return err
}

// DeleteDevice unregister device from receiving notification
func (service *NotificationService) DeleteDevice(token string) error {
	err := service.fcmRepo.DeleteToken(token)
	return err
}

// TODO: current time to expire is ... ?

// GetUserDevices return array of tokens of user devices, excluding expired one
func (service *NotificationService) GetUserDevices(userID string) ([]string, error) {
	tokenIDs, err := service.fmcUserRepo.GetUserTokens(userID)
	if err != nil {
		return nil, err
	}

	tokens, err := service.fcmRepo.GetTokensByIDs(tokenIDs)
	nonExpiredTokens := make([]string, 0)

	now := time.Now()
	for _, tok := range tokens {
		if now.Sub(tok.LastUpdated) <= 24*time.Hour {
			nonExpiredTokens = append(nonExpiredTokens, tok.Token)
		}
	}
	return nonExpiredTokens, nil
}
