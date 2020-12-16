package repository

import (
	"backendSenior/domain/model"
)

// RoomRepository defines interface for room repo
type RoomRepository interface {
	GetAllRooms() ([]model.Room, error)
	GetRoomByID(roomID string) (model.Room, error)

	AddRoom(room model.Room) (string, error)
	UpdateRoom(roomID string, room model.Room) error
	DeleteRoomByID(roomID string) error
}
