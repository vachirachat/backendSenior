package delegate

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/utills"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/go-resty/resty/v2"
	"net/url"
	"proxySenior/utils"
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
	clnt                 *resty.Client
}

func (repo *DelegateRoomUserRepository) AddAdminsToRoom(roomID bson.ObjectId, userIDs []bson.ObjectId) (err error) {
	panic("not supported")
}

func (repo *DelegateRoomUserRepository) RemoveAdminsFromRoom(roomID bson.ObjectId, userIDs []bson.ObjectId) (err error) {
	panic("not supported")
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
		clnt:                 resty.New(),
	}
	return repo
}

// GetUserRooms get user's room from backend API
func (repo *DelegateRoomUserRepository) GetUserRooms(userID string) (roomIDs []string, err error) {
	repo.lock.RLock()
	fetchTime := repo.lastFetchUsertoRooms[userID]
	rooms := repo.userToRooms[userID]
	repo.lock.RUnlock()

	if fetchTime.IsZero() || time.Now().Sub(fetchTime) > repo.ttl {
		url := url.URL{
			Scheme: "http",
			Host:   repo.controllerOrigin,
			Path:   "/api/v1/user/byid/" + userID,
		}

		var resOk model.User
		if res, err := repo.clnt.R().SetHeader("Authorization", utils.AuthHeader()).SetResult(&resOk).Get(url.String()); err != nil {
			return nil, fmt.Errorf("get user rooms: request error: %s", err)
		} else if res.IsError() {
			return nil, fmt.Errorf("get user rooms: server replied with status: %d", res.StatusCode())
		}

		repo.lock.Lock()
		defer repo.lock.Unlock()

		repo.userToRooms[userID] = utills.ToStringArr(resOk.Room)
		repo.lastFetchUsertoRooms[userID] = time.Now()

		return repo.userToRooms[userID], nil
	}
	return rooms, nil
}

// GetRoomUsers get room's users from backend API
func (repo *DelegateRoomUserRepository) GetRoomUsers(roomID string) (userIDs []string, err error) {
	repo.lock.RLock()
	fetchTime := repo.lastFetchRoomToUsers[roomID]
	users := repo.roomToUsers[roomID]
	repo.lock.RUnlock()

	if fetchTime.IsZero() || time.Now().Sub(fetchTime) > repo.ttl {
		url := url.URL{
			Scheme: "http",
			Host:   repo.controllerOrigin,
			Path:   "/api/v1/room/id/" + roomID + "/member",
		}
		var resOk struct {
			Members []string `json:"members"`
		}
		if res, err := repo.clnt.R().SetHeader("Authorization", utils.AuthHeader()).SetResult(&resOk).Get(url.String()); err != nil {
			return nil, fmt.Errorf("get room users: request error: %s", err)
		} else if res.IsError() {
			return nil, fmt.Errorf("get room users: server replied with status: %d", res.StatusCode())
		}

		repo.lock.Lock()
		defer repo.lock.Unlock()

		repo.roomToUsers[roomID] = resOk.Members
		repo.lastFetchRoomToUsers[roomID] = time.Now()

		return repo.roomToUsers[roomID], nil
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
