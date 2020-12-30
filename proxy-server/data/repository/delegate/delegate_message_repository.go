package delegate

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// DelegateMessageRepository is repository for getting message from controller
type DelegateMessageRepository struct {
	controllerOrigin string // origin is hostname and port
}

func NewDelegateMessageRepository(origin string) *DelegateMessageRepository {
	return &DelegateMessageRepository{
		controllerOrigin: origin,
	}
}

var _ repository.MessageRepository = (*DelegateMessageRepository)(nil)

func (repo *DelegateMessageRepository) GetAllMessages(timeRange *model.TimeRange) ([]model.Message, error) {
	url := url.URL{
		Scheme: "http",
		Host:   repo.controllerOrigin,
		Path:   "/api/v1/message",
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
		Messages []model.Message `json:"messages"`
	}
	err = json.Unmarshal(body, &resOk)
	if err != nil {
		return nil, err
	}

	return resOk.Messages, nil
}

func (repo *DelegateMessageRepository) GetMessagesByRoom(roomID string, timeRange *model.TimeRange) ([]model.Message, error) {
	url := url.URL{
		Scheme:   "http",
		Host:     repo.controllerOrigin,
		Path:     "/api/v1/message",
		RawQuery: "roomId=" + roomID,
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
		Messages []model.Message `json:"messages"`
	}
	err = json.Unmarshal(body, &resOk)
	if err != nil {
		return nil, err
	}

	return resOk.Messages, nil
}

func (repo *DelegateMessageRepository) GetMessageByID(messageID string) (model.Message, error) {
	panic("not available")
}
func (repo *DelegateMessageRepository) AddMessage(message model.Message) (string, error) {
	panic("not available")
}
func (repo *DelegateMessageRepository) DeleteMessageByID(userID string) error {
	panic("not available")
}
