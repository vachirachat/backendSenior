package service

import (
	"backendSenior/domain/interface/repository"

	"backendSenior/domain/model"
)

type RoomService struct {
	roomRepository      repository.RoomRepository
	roomUserRepository  repository.RoomUserRepository
	roomProxyRepository repository.RoomUserRepository
}

// NewRoomService create new room service instance
func NewRoomService(roomRepo repository.RoomRepository, roomUser repository.RoomUserRepository, roomProxy repository.RoomUserRepository) *RoomService {
	return &RoomService{
		roomRepository:      roomRepo,
		roomUserRepository:  roomUser,
		roomProxyRepository: roomProxy,
	}
}

// GetAllRooms get all rooms from all user
func (service *RoomService) GetAllRooms() ([]model.Room, error) {
	rooms, err := service.roomRepository.GetAllRooms()
	return rooms, err
}

// GetRoomByID get room by Id
func (service *RoomService) GetRoomByID(roomID string) (model.Room, error) {
	room, err := service.roomRepository.GetRoomByID(roomID)
	return room, err
}

// AddRoom insert room into database and return id of newly inserted room
// The created room will always be empty (need to invite as separate request)
func (service *RoomService) AddRoom(room model.Room) (string, error) {
	roomID, err := service.roomRepository.AddRoom(room)
	return roomID, err
}

// EditRoomName change name of room
// todo this should pass only room name
func (service *RoomService) EditRoomName(roomID string, room model.Room) error {
	err := service.roomRepository.UpdateRoom(roomID, room)
	return err
}

// DeleteRoomByID delete a room by id
func (service *RoomService) DeleteRoomByID(roomID string) error {
	err := service.roomRepository.DeleteRoomByID(roomID)
	return err
}

// Match with Socket-structure

// AddMembersToRoom add members to room
func (service *RoomService) AddMembersToRoom(roomID string, userList []string) error {
	err := service.roomUserRepository.AddUsersToRoom(roomID, userList)
	return err
}

// DeleteMemberFromRoom removes members from room
func (service *RoomService) DeleteMemberFromRoom(roomID string, userList []string) error {
	err := service.roomUserRepository.RemoveUsersFromRoom(roomID, userList)
	return err
}

// GetRoomMembers return list of members in rooms
func (service *RoomService) GetRoomMembers(roomID string) ([]string, error) {
	members, err := service.roomUserRepository.GetRoomUsers(roomID)
	return members, err
}

// -- proxy management part

// AddProxiesToRoom add proxies to room
func (service *RoomService) AddProxiesToRoom(roomID string, userList []string) error {
	err := service.roomProxyRepository.AddUsersToRoom(roomID, userList)
	return err
}

// DeleteProxiesFromRoom removes proxies from room
func (service *RoomService) DeleteProxiesFromRoom(roomID string, userList []string) error {
	err := service.roomProxyRepository.RemoveUsersFromRoom(roomID, userList)
	return err
}

// GetRoomProxies return list of proxies in rooms
func (service *RoomService) GetRoomProxies(roomID string) ([]string, error) {
	members, err := service.roomProxyRepository.GetRoomUsers(roomID)
	return members, err
}
