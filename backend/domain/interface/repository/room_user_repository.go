package repository

// RoomUserRepository is interface for repository managing room/user relation
type RoomUserRepository interface {
	GetUserRooms(userID string) (roomIDs []string, err error)
	GetRoomUsers(roomID string) (userIDs []string, err error)
	AddUsersToRoom(roomID string, userIDs []string) (err error)
	RemoveUsersFromRoom(roomID string, userIDs []string) (err error)
}
