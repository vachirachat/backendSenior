package repository

import (
	"backendSenior/domain/model"
)

type TokenRepository interface {
	CountToken(filter interface{}) (int, error)
	InsertToken(token model.TokenDB) error
	RemoveTokens(filter interface{}) (int, error)
}
