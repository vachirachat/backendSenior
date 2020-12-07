package service

import (
	"backendSenior/domain/interface/repository"

	"backendSenior/domain/model"

	"github.com/globalsign/mgo/bson"
)

type RoomService struct {
	roomRepository repository.RoomRepository
}

// NewRoomService create new room service instance
func NewRoomService(roomRepo repository.RoomRepository) *RoomService {
	return &RoomService{
		roomRepository: roomRepo,
	}
}

// GetAllRooms get all rooms from all user
func (service *RoomService) GetAllRooms() ([]model.Room, error) {
	rooms, err := service.roomRepository.GetAllRooms()
	return rooms, err
}

// GetRoomByID get room by Id
func (service *RoomService) GetRoomByID(roomID bson.ObjectId) (model.Room, error) {
	room, err := service.roomRepository.GetRoomByID(roomID)
	return room, err
}

// AddRoom insert room into database and return id of newly inserted room
func (service *RoomService) AddRoom(room model.Room) (bson.ObjectId, error) {
	roomID, err := service.roomRepository.AddRoom(room)
	return roomID, err
}

// EditRoomName change name of room
// todo this should pass only room name
func (service *RoomService) EditRoomName(roomID bson.ObjectId, room model.Room) error {
	err := service.roomRepository.EditRoomName(room.RoomID, room)
	return err
}

// DeleteRoomByID delete a room by id
func (service *RoomService) DeleteRoomByID(roomID bson.ObjectId) error {
	err := service.roomRepository.DeleteRoomByID(roomID)
	return err
}

// Match with Socket-structure

// AddMembersToRoom add members to room
func (service *RoomService) AddMembersToRoom(roomID bson.ObjectId, userList []bson.ObjectId) error {
	err := service.roomRepository.AddMemberToRoom(roomID, userList)
	return err
}

// DeleteMemberFromRoom removes a member from room
func (service *RoomService) DeleteMemberFromRoom(roomID bson.ObjectId, userID bson.ObjectId) error {
	err := service.roomRepository.DeleteMemberFromRoom(userID, roomID)
	return err
}
