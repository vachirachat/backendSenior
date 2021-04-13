package dto

import (
	"backendSenior/domain/model"
	"time"
)

type FCMTokenDto struct {
	Token      string `json:"token" validate:"required,gt=0"`
	DeviceName string `json:"deviceName" validate:"gt=0"`
}

func (d *FCMTokenDto) ToFCMToken() model.FCMToken {
	return model.FCMToken{
		Token:       d.Token,
		DeviceName:  d.DeviceName,
		LastUpdated: time.Now(),
	}
}
