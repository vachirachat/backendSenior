package service

import (
	"backendSenior/domain/interface/repository"

	"backendSenior/domain/model"

	"github.com/globalsign/mgo/bson"
)

type RoomService struct {
	roomRepository      repository.RoomRepository
	roomUserRepository  repository.RoomUserRepository
	roomProxyRepository repository.RoomProxyRepository
}

// NewRoomService create new room service instance
func NewRoomService(roomRepo repository.RoomRepository, roomUser repository.RoomUserRepository, roomProxy repository.RoomProxyRepository) *RoomService {
	return &RoomService{
		roomRepository:      roomRepo,
		roomUserRepository:  roomUser,
		roomProxyRepository: roomProxy,
	}
}

// GetAllRooms get all rooms from all user
func (s *RoomService) GetAllRooms() ([]model.Room, error) {
	rooms, err := s.roomRepository.GetAllRooms()
	return rooms, err
}

// GetRoomByID get room by Id
func (s *RoomService) GetRoomByID(roomID string) (model.Room, error) {
	room, err := s.roomRepository.GetRoomByID(roomID)
	return room, err
}

// GetRoomsByIDs return rooms by specifying array of ID
func (s *RoomService) GetRoomsByIDs(roomIDs []string) ([]model.Room, error) {
	rooms, err := s.roomRepository.GetRoomsByIDs(roomIDs)
	return rooms, err
}

// GetUserRooms return rooms of user, it get roomIds of user by mapping then query rooms by ID
func (s *RoomService) GetUserRooms(userID string) ([]model.Room, error) {
	roomIDs, err := s.roomUserRepository.GetUserRooms(userID)
	if err != nil {
		return nil, err
	}
	rooms, err := s.roomRepository.GetRoomsByIDs(roomIDs)
	return rooms, err
}

// AddRoom insert room into database and return id of newly inserted room
// The created room will always be empty (need to invite as separate request)
func (s *RoomService) AddRoom(room model.Room) (string, error) {
	room.ListUser = []bson.ObjectId{}
	room.ListProxy = []bson.ObjectId{}
	roomID, err := s.roomRepository.AddRoom(room)
	return roomID, err
}

// EditRoomName change name of room
// todo this should pass only room name
func (s *RoomService) EditRoomName(roomID string, room model.Room) error {
	err := s.roomRepository.UpdateRoom(roomID, room)
	return err
}

// DeleteRoomByID delete a room by id
func (s *RoomService) DeleteRoomByID(roomID string) error {
	err := s.roomRepository.DeleteRoomByID(roomID)
	return err
}

// Match with Socket-structure

// AddMembersToRoom add members to room
func (s *RoomService) AddMembersToRoom(roomID string, userList []string) error {
	err := s.roomUserRepository.AddUsersToRoom(roomID, userList)
	return err
}

// DeleteMemberFromRoom removes members from room
func (s *RoomService) DeleteMemberFromRoom(roomID string, userList []string) error {
	err := s.roomUserRepository.RemoveUsersFromRoom(roomID, userList)
	return err
}

// GetRoomMemberIDs return list of members in rooms (as ID)
func (s *RoomService) GetRoomMemberIDs(roomID string) ([]string, error) {
	members, err := s.roomUserRepository.GetRoomUsers(roomID)
	return members, err
}

// -- proxy management part

// AddProxiesToRoom add proxies to room
func (s *RoomService) AddProxiesToRoom(roomID string, userList []string) error {
	err := s.roomProxyRepository.AddProxiesToRoom(roomID, userList)
	return err
}

// DeleteProxiesFromRoom removes proxies from room
func (s *RoomService) DeleteProxiesFromRoom(roomID string, userList []string) error {
	err := s.roomProxyRepository.RemoveProxiesFromRoom(roomID, userList)
	return err
}

// GetRoomProxyIDs return list of proxies in rooms as ID
func (s *RoomService) GetRoomProxyIDs(roomID string) ([]string, error) {
	members, err := s.roomProxyRepository.GetRoomProxies(roomID)
	return members, err
}

// Proxy master no longer fixed

// // GetRoomMasterProxy return master proxy of the room
// func (s *RoomService) GetRoomMasterProxy(roomID string) (masterProxy model.Proxy, err error) {
// 	masterProxy, err = s.roomProxyRepository.GetRoomMasterProxy(roomID)
// 	return
// }

// // SetRoomMasterProxy set proxy that is master of the room
// func (s *RoomService) SetRoomMasterProxy(roomID string, masterProxyID string) error {
// 	err := s.roomProxyRepository.SetRoomMasterProxy(roomID, masterProxyID)
// 	return err
// }

// // GetProxyMasterRooms get rooms of which proxy is master of
// func (s *RoomService) GetProxyMasterRooms(proxyID string) ([]string, error) {
// 	roomIDs, err := s.roomProxyRepository.GetProxyMasterRooms(proxyID)
// 	return roomIDs, err
// }
