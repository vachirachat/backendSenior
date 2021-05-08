package delegate

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/url"
	"proxySenior/utils"
)

// DelegateMessageRepository is repository for getting message from controller
type DelegateMessageRepository struct {
	controllerOrigin string // origin is hostname and port
	clnt             *resty.Client
}

func NewDelegateMessageRepository(origin string) *DelegateMessageRepository {
	return &DelegateMessageRepository{
		controllerOrigin: origin,
		clnt:             resty.New(),
	}
}

var _ repository.MessageRepository = (*DelegateMessageRepository)(nil)

func (repo *DelegateMessageRepository) GetAllMessages(timeRange *model.TimeRange) ([]model.Message, error) {
	url := url.URL{
		Scheme: "http",
		Host:   repo.controllerOrigin,
		Path:   "/api/v1/message",
	}
	var resOk struct {
		Messages []model.Message `json:"messages"`
	}
	if res, err := repo.clnt.R().SetHeader("Authorization", utils.AuthHeader()).SetResult(&resOk).Get(url.String()); err != nil {
		return nil, fmt.Errorf("error in request: %s", err)
	} else if res.IsError() {
		return nil, fmt.Errorf("error in request: server returned status code %d", res.StatusCode())
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
	var resOk struct {
		Messages []model.Message `json:"messages"`
	}
	if res, err := repo.clnt.R().SetHeader("Authorization", utils.AuthHeader()).SetResult(&resOk).Get(url.String()); err != nil {
		return nil, fmt.Errorf("error in request: %s", err)
	} else if res.IsError() {
		return nil, fmt.Errorf("error in request: server returned status code %d", res.StatusCode())
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
