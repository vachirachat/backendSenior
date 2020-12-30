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
	"sync"
	"time"
)

type DelegateRoomUserRepository struct {
	roomToUsers          map[string][]string
	userToRooms          map[string][]string
	lastFetchRoomToUsers map[string]time.Time
	lastFetchUsertoRooms map[string]time.Time
	controllerOrigin     string        // origin is hostname and port
	ttl                  time.Duration // cache duration
	lock                 sync.RWMutex
}

var _ repository.RoomUserRepository = (*DelegateRoomUserRepository)(nil)

func NewDelegateRoomUserRepository(controllerOrigin string) *DelegateRoomUserRepository {
	repo := &DelegateRoomUserRepository{
		roomToUsers:          make(map[string][]string),
		userToRooms:          make(map[string][]string),
		lastFetchRoomToUsers: make(map[string]time.Time),
		lastFetchUsertoRooms: make(map[string]time.Time),
		controllerOrigin:     controllerOrigin,
		ttl:                  60 * time.Second,
		lock:                 sync.RWMutex{},
	}
	return repo
}

// GetUserRooms get user's room from backend API
func (repo *DelegateRoomUserRepository) GetUserRooms(userID string) (roomIDs []string, err error) {
	repo.lock.RLock()
	fetchTime := repo.lastFetchUsertoRooms[userID]
	rooms := repo.userToRooms[userID]
	repo.lock.RUnlock()

	if time.Now().Sub(fetchTime) > repo.ttl {
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

		repo.lock.Lock()
		defer repo.lock.Unlock()

		repo.userToRooms[userID] = utills.ToStringArr(resOk.Room)
		repo.lastFetchUsertoRooms[userID] = time.Now()
	}
	return rooms, nil
}

// GetRoomUsers get room's users from backend API
func (repo *DelegateRoomUserRepository) GetRoomUsers(roomID string) (userIDs []string, err error) {
	repo.lock.RLock()
	fetchTime := repo.lastFetchUsertoRooms[roomID]
	users := repo.userToRooms[roomID]
	repo.lock.RUnlock()

	if time.Now().Sub(fetchTime) > repo.ttl {
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

		repo.lock.Lock()
		defer repo.lock.Unlock()

		repo.roomToUsers[roomID] = resOk.Members
		repo.lastFetchRoomToUsers[roomID] = time.Now()
	}
	return users, nil
}

// AddUsersToRoom is used for updating cached users, it DOES NOT update database
func (repo *DelegateRoomUserRepository) AddUsersToRoom(roomID string, userIDs []string) (err error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	repo.roomToUsers[roomID] = append(repo.roomToUsers[roomID], userIDs...)
	for _, userID := range userIDs {
		repo.userToRooms[userID] = append(repo.userToRooms[userID], roomID)
	}

	return nil
}

// RemoveUsersFromRoom is used for updating cached users, it DOES NOT update database
func (repo *DelegateRoomUserRepository) RemoveUsersFromRoom(roomID string, userIDs []string) (err error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	repo.roomToUsers[roomID], _ = utills.ArrStringRemoveMatched(repo.roomToUsers[roomID], userIDs)
	for _, userID := range userIDs {
		repo.userToRooms[userID] = utills.RemoveFormListString(repo.userToRooms[userID], roomID)
	}

	return nil
}
