package delegate

import (
	"backendSenior/domain/model"
	"fmt"
	"github.com/go-resty/resty/v2"
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

	var masterProxy model.Proxy
	if res, err := a.c.R().SetHeader("Authorization", utils.AuthHeader()).SetResult(&masterProxy).Get(u.String()); err != nil {
		return model.Proxy{}, fmt.Errorf("get room master proxy: request error: %s", err)
	} else if res.IsError() {
		return model.Proxy{}, fmt.Errorf("get room master proxy: server returned status code %d", res.StatusCode())
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

	if _, err := a.c.R().
		SetHeader("Authorization", utils.AuthHeader()).
		SetResult(&r).Get(u.String()); err != nil {
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
