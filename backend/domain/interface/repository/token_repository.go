package repository

import (
	"backendSenior/domain/model"
)

type TokenRepository interface {
	VerifyDBToken(userid string, accessToken string) (string, error)
	AddToken(userid string, accessToken string) error
	RemoveToken(userid string) error
	GetAllToken() ([]model.TokenDB, error)
}
