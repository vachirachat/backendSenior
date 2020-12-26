package repository

import "backendSenior/domain/model"

// FCMTokenRepository define interface for managing FCM tokens
type FCMTokenRepository interface {
	GetAllTokens() ([]model.FCMToken, error)
	GetTokenByID(token string) (model.FCMToken, error)
	GetTokensByIDs(tokens []string) ([]model.FCMToken, error)
	AddToken(token model.FCMToken) error
	DeleteToken(token string) error
	UpdateToken(token string, update model.FCMToken) error
}

// FCMUserRepository manage relation between token and user
type FCMUserRepository interface {
	// Get all token owned by user
	GetUserTokens(userID string) (tokenIDs []string, err error)
	// Get owner of token
	GetTokenOwner(token string) (tokenID string, err error)
	// Associate token to user
	AddUserToken(userID string, tokenID string) (err error)
	// Disassociate token from user
	DeleteUserToken(userID string, tokenID string) (err error)
}
