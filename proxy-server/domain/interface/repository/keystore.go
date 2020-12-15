package repository

import "backendSenior/domain/model"

type Keystore interface {
	GetKeyForMessage(message *model.Message) (key string, err error)
}
