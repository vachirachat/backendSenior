package authorization

import "backendSenior/domain/model"

// AuthorizationService defines generic interface to verify user accesss
type AuthorizationService interface {
	IsAuthorized(userDetail model.UserDetail, resouceID string, action string) (ok bool, err error)
	// TODO: in the future add some method for managing access
}
