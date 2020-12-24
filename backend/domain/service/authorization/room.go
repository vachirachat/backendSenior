package authorization

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
)

type RoomAuthorization struct {
	roomRepo *repository.RoomRepository
}

func NewRoomAuthorization(roomRepo *repository.RoomRepository) *RoomAuthorization {
	return &RoomAuthorization{
		roomRepo: roomRepo,
	}
}

var _ (AuthorizationService) = (*RoomAuthorization)(nil)

// IsAuthorized return whether user is permiited to do `action` on room `roomID`
func (auth *RoomAuthorization) IsAuthorized(userDetail model.UserDetail, roomID string, action string) (ok bool, err error) {
	return true, nil
}
