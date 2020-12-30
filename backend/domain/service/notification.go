package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/globalsign/mgo/bson"
)

type NotificationService struct {
	fcmRepo     repository.FCMTokenRepository
	fcmUserRepo repository.FCMUserRepository
	fcmClient   *messaging.Client
	lastSeen    map[string]time.Time
	isOnline    map[string]bool
	seenLock    sync.RWMutex
	onlineLock  sync.RWMutex
}

func NewNotificationService(fcmRepo repository.FCMTokenRepository, fcmUserRepo repository.FCMUserRepository, fcmClient *messaging.Client) *NotificationService {
	return &NotificationService{
		fcmRepo:     fcmRepo,
		fcmUserRepo: fcmUserRepo,
		fcmClient:   fcmClient,
		lastSeen:    make(map[string]time.Time),
		isOnline:    make(map[string]bool),
		seenLock:    sync.RWMutex{},
		onlineLock:  sync.RWMutex{},
	}
}

// RegisterDevice register user device (specified by token) for receiving notificiation for that user
func (service *NotificationService) RegisterDevice(userID string, deviceToken string, deviceName string) error {
	_, err := service.fcmRepo.GetTokenByID(deviceToken)
	// found token
	if err == nil {
		return errors.New("token already registered")
	}

	// some other error
	if err != nil && err.Error() != "not found" {
		return err
	}

	err = service.fcmRepo.AddToken(model.FCMToken{
		Token:       deviceToken,
		UserID:      bson.ObjectIdHex(userID),
		LastUpdated: time.Now(),
		DeviceName:  deviceName,
	})

	if err != nil {
		return err
	}

	err = service.fcmUserRepo.AddUserToken(userID, deviceToken)

	return err
}

// RefreshDevice refresh last update timestamp of a token, to prevent it from expiring
func (service *NotificationService) RefreshDevice(deviceToken string) error {
	err := service.fcmRepo.UpdateToken(deviceToken, model.FCMToken{
		LastUpdated: time.Now(),
	})

	return err
}

// DeleteDevice unregister device from receiving notification
func (service *NotificationService) DeleteDevice(deviceToken string) error {
	existToken, err := service.fcmRepo.GetTokenByID(deviceToken)
	if err != nil {
		return err
	}

	err = service.fcmRepo.DeleteToken(deviceToken)

	if err != nil {
		return err
	}

	err = service.fcmUserRepo.DeleteUserToken(existToken.UserID.Hex(), deviceToken)

	return err
}

// GetUserTokens return array of tokens of user devices
func (service *NotificationService) GetUserTokens(userID string) ([]model.FCMToken, error) {
	tokenIDs, err := service.fcmUserRepo.GetUserTokens(userID)
	if err != nil {
		return nil, err
	}

	tokens, err := service.fcmRepo.GetTokensByIDs(tokenIDs)
	return tokens, err
}

// SetLastSeenTime sets last seen time of device (token)
func (service *NotificationService) SetLastSeenTime(token string, time time.Time) {
	service.seenLock.Lock()
	defer service.seenLock.Unlock()

	service.lastSeen[token] = time
}

// SetOnlineStatus sets last seen time of device (token)
func (service *NotificationService) SetOnlineStatus(token string, status bool) {
	service.onlineLock.Lock()
	defer service.onlineLock.Unlock()
	service.isOnline[token] = status
}

// GetLastSeenTime returns last seen time of device
func (service *NotificationService) GetLastSeenTime(token string) time.Time {
	service.seenLock.RLock()
	defer service.seenLock.RUnlock()
	return service.lastSeen[token]
}

// GetOnlineStatus returns last seen time of device
func (service *NotificationService) GetOnlineStatus(token string) bool {
	service.onlineLock.RLock()
	defer service.onlineLock.RUnlock()
	return service.isOnline[token]
}

// GetTokenByID returns token by ID
func (service *NotificationService) GetTokenByID(token string) (model.FCMToken, error) {
	foundToken, err := service.fcmRepo.GetTokenByID(token)
	return foundToken, err
}

type SendError struct {
	BatchResponse *messaging.BatchResponse
}

func (err *SendError) Error() string {
	return fmt.Sprintf("error sending %d/%d messages", err.BatchResponse.FailureCount, err.BatchResponse.FailureCount+err.BatchResponse.SuccessCount)
}

// SendNotifications sends notification to all of devices
// It returns number of success repsonse and error if any of them send unsuccessfully
func (service *NotificationService) SendNotifications(deviceTokens []string, notification *model.Notification) (int, error) {
	resp, err := service.fcmClient.SendMulticast(context.Background(), &messaging.MulticastMessage{
		Tokens: deviceTokens,
		Data:   notification.Data,
		Notification: &messaging.Notification{
			Title:    notification.Title,
			Body:     notification.Body,
			ImageURL: notification.ImageURL,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
	})

	if resp == nil {
		return 0, err
	}

	if resp.FailureCount == 0 {
		return resp.SuccessCount, nil
	}

	return resp.SuccessCount, &SendError{
		BatchResponse: resp,
	}

}
