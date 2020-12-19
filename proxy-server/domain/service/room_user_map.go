package service

import (
	"backendSenior/domain/interface/repository"
)

type RoomUserMap struct {
	roomUserRepo repository.RoomUserRepository
}

func NewRoomUserMap(roomUserRepo repository.RoomUserRepository) *RoomUserMap {
	return &RoomUserMap{
		roomUserRepo: roomUserRepo,
	}
}

func (m *RoomUserMap) GetRoomUsers(roomID string) ([]string, error) {
	return m.roomUserRepo.GetRoomUsers(roomID)
}

func (m *RoomUserMap) GetUserRooms(userID string) ([]string, error) {
	return m.roomUserRepo.GetUserRooms(userID)
}

func (m *RoomUserMap) IsUserInRoom(userID string, roomID string) (bool, error) {
	rooms, err := m.roomUserRepo.GetUserRooms(userID)
	if err != nil {
		return false, err
	}
	for _, u := range rooms {
		if u == roomID {
			return true, nil
		}
	}
	return false, nil
}
