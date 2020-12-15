package repository

// import (
// 	"backendSenior/domain/interface/repository"
// 	"net/http"
// 	"time"
// )

// type DelegateRoomUserRepository struct {
// 	roomToUsers          map[string][]string
// 	userToRooms          map[string][]string
// 	lastFetchRoomToUsers map[string]time.Time
// 	lastFetchUsertoRooms map[string]time.Time
// 	controllerOrigin     string        // origin is hostname and port
// 	ttl                  time.Duration // cache duration
// }

// var _ repository.RoomUserRepository = (*DelegateRoomUserRepository)(nil)

// func NewDelegateRoomUserRepository(controllerOrigin string) *DelegateRoomUserRepository {
// 	repo := &DelegateRoomUserRepository{
// 		roomToUsers:      make(map[string][]string),
// 		userToRooms:      make(map[string][]string),
// 		controllerOrigin: controllerOrigin,
// 		ttl:              10 * time.Second,
// 	}
// }

// func (repo *DelegateRoomUserRepository) GetUserRooms(userID string) (roomIDs []string, err error) {
// 	if time.Now().Sub(repo.lastFetchUsertoRooms[userID]) > repo.ttl {
// 		// TODO
// 		http.Get("http://" + repo.controllerOrigin + "/api/v1/user/rooms")
// 	}
// }
// func (repo *DelegateRoomUserRepository) GetRoomUsers(roomID string) (userIDs []string, err error) {

// }
// func (repo *DelegateRoomUserRepository) AddUsersToRoom(roomID string, userIDs []string) (err error) {

// }
// func (repo *DelegateRoomUserRepository) RemoveUsersFromRoom(roomID string, userIDs []string) (err error) {

// }
