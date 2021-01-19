package delegate

import (
	"backendSenior/domain/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"proxySenior/domain/interface/repository"
)

type RoomProxyAPI struct {
	origin string // host:port of controller
}

var _ repository.ProxyMasterAPI = (*RoomProxyAPI)(nil)

func NewRoomProxyAPI(controller string) *RoomProxyAPI {
	return &RoomProxyAPI{
		origin: controller,
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
	res, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("server return with non OK status: %d\nbody:%s", res.StatusCode, body)
	}

	var r struct { // see backend/proxy_route_handler
		RoomIDs []string `json:"roomIds"`
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return r.RoomIDs, nil
}
