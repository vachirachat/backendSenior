package delegate

import (
	"backendSenior/domain/model"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"proxySenior/domain/interface/repository"
	"proxySenior/utils"
)

type RoomProxyAPI struct {
	origin string // host:port of controller
	c      *resty.Client
}

var _ repository.ProxyMasterAPI = (*RoomProxyAPI)(nil)

func NewRoomProxyAPI(controller string) *RoomProxyAPI {
	return &RoomProxyAPI{
		origin: controller,
		c:      resty.New(),
	}
}

// GetRoomMasterProxy make request to return proxy that is master of the room
func (a *RoomProxyAPI) GetRoomMasterProxy(roomID string) (model.Proxy, error) {
	u := url.URL{
		Scheme: "http",
		Host:   a.origin,
		Path:   "/api/v1/key/master-proxy/" + roomID,
	}
	res, err := http.Get(u.String())
	if err != nil {
		return model.Proxy{}, fmt.Errorf("error making request: %v", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return model.Proxy{}, fmt.Errorf("server return with non OK status: %d\nbody:%s", res.StatusCode, body)
	}

	var masterProxy model.Proxy
	err = json.Unmarshal(body, &masterProxy)
	if err != nil {
		return model.Proxy{}, fmt.Errorf("error decoding response: %v", err)
	}

	return masterProxy, nil
}

// GetProxyMasterRooms make request to get master roomIDs
func (a *RoomProxyAPI) GetProxyMasterRooms(proxyID string) ([]string, error) {
	u := url.URL{
		Scheme: "http",
		Host:   a.origin,
		Path:   "/api/v1/proxy/" + proxyID + "/master-rooms",
	}

	var r struct { // see backend/proxy_route_handler
		RoomIDs []string `json:"roomIds"`
	}
	err := utils.HTTPGet(u.String(), &r)
	if err != nil {
		return nil, fmt.Errorf("get proxy master rooms: %w", err)
	}

	return r.RoomIDs, nil
}

// GetProxyByID call controller API for getting proxy by ID
func (a *RoomProxyAPI) GetProxyByID(proxyID string) (model.Proxy, error) {
	u := url.URL{
		Scheme: "http",
		Host:   a.origin,
		Path:   "/api/v1/proxy/" + proxyID,
	}

	var p model.Proxy
	_, err := a.c.R().
		SetResult(&p).
		SetHeader("Authorization", utils.AuthHeader()).
		Get(u.String())
	if err != nil {
		return model.Proxy{}, fmt.Errorf("get proxy by id: %w", err)
	}

	return p, nil
}
