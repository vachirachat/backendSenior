package delegate

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/utills"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type DelegateRoomUserRepository struct {
	roomToUsers          map[string][]string
	userToRooms          map[string][]string
	lastFetchRoomToUsers map[string]time.Time
	lastFetchUsertoRooms map[string]time.Time
	controllerOrigin     string        // origin is hostname and port
	ttl                  time.Duration // cache duration
}

var _ repository.RoomUserRepository = (*DelegateRoomUserRepository)(nil)

func NewDelegateRoomUserRepository(controllerOrigin string) *DelegateRoomUserRepository {
	repo := &DelegateRoomUserRepository{
		roomToUsers:          make(map[string][]string),
		userToRooms:          make(map[string][]string),
		lastFetchRoomToUsers: make(map[string]time.Time),
		lastFetchUsertoRooms: make(map[string]time.Time),
		controllerOrigin:     controllerOrigin,
		ttl:                  10 * time.Second,
	}
	return repo
}

// GetUserRooms get user's room from backend API
func (repo *DelegateRoomUserRepository) GetUserRooms(userID string) (roomIDs []string, err error) {
	if time.Now().Sub(repo.lastFetchUsertoRooms[userID]) > repo.ttl {
		// TODO
		url := url.URL{
			Scheme: "http",
			Host:   repo.controllerOrigin,
			Path:   "/api/v1/user/byid/" + userID,
		}

		http.Get(url.String())
		res, err := http.Get(url.String())
		if err != nil {
			return nil, err
		} else if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("server response with status " + res.Status)
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		var resOk model.User
		err = json.Unmarshal(body, &resOk)
		if err != nil {
			return nil, err
		}

		repo.userToRooms[userID] = utills.ToStringArr(resOk.Room)
		repo.lastFetchUsertoRooms[userID] = time.Now()
	}
	return repo.userToRooms[userID], nil
}

// GetRoomUsers get room's users from backend API
func (repo *DelegateRoomUserRepository) GetRoomUsers(roomID string) (userIDs []string, err error) {
	if time.Now().Sub(repo.lastFetchRoomToUsers[roomID]) > repo.ttl {
		url := url.URL{
			Scheme: "http",
			Host:   repo.controllerOrigin,
			Path:   "/api/v1/room/" + roomID + "/member",
		}
		res, err := http.Get(url.String())
		if err != nil {
			return nil, err
		} else if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("server response with status " + res.Status)
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		var resOk struct {
			Members []string `json:"members"`
		}
		err = json.Unmarshal(body, &resOk)
		if err != nil {
			return nil, err
		}

		repo.roomToUsers[roomID] = resOk.Members
		repo.lastFetchRoomToUsers[roomID] = time.Now()
	}
	return repo.roomToUsers[roomID], nil
}
func (repo *DelegateRoomUserRepository) AddUsersToRoom(roomID string, userIDs []string) (err error) {
	panic("Not Allowed")
}
func (repo *DelegateRoomUserRepository) RemoveUsersFromRoom(roomID string, userIDs []string) (err error) {
	panic("Not Allowed")
}
