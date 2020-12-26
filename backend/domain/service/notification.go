package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"context"
	"errors"
	"fmt"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/globalsign/mgo/bson"
)

type NotificationService struct {
	fcmRepo     repository.FCMTokenRepository
	fcmUserRepo repository.FCMUserRepository
	fcmClient   *messaging.Client
}

func NewNotificationService(fcmRepo repository.FCMTokenRepository, fcmUserRepo repository.FCMUserRepository, fcmClient *messaging.Client) *NotificationService {
	return &NotificationService{
		fcmRepo:     fcmRepo,
		fcmUserRepo: fcmUserRepo,
		fcmClient:   fcmClient,
	}
}

// RegisterDevice register user device (specified by token) for receiving notificiation for that user
func (service *NotificationService) RegisterDevice(userID string, deviceToken string) error {
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

// GetUserDevices return array of tokens of user devices
func (service *NotificationService) GetUserTokens(userID string) ([]model.FCMToken, error) {
	tokenIDs, err := service.fcmUserRepo.GetUserTokens(userID)
	if err != nil {
		return nil, err
	}

	tokens, err := service.fcmRepo.GetTokensByIDs(tokenIDs)
	return tokens, err
	// nonExpiredTokens := make([]string, 0)

	// now := time.Now()
	// for _, tok := range tokens {
	// 	if now.Sub(tok.LastUpdated) <= 24*time.Hour {
	// 		nonExpiredTokens = append(nonExpiredTokens, tok.Token)
	// 	}
	// }
	// return nonExpiredTokens, nil
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
