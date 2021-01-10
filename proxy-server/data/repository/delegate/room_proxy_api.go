package delegate

import (
	"backendSenior/domain/model"
	"net/url"
	"proxySenior/domain/interface/repository"
)

type RoomProxyAPI struct {
	origin string // host:port of controller
}

var _ repository.ProxyMasterRepo = (*RoomProxyAPI)(nil)

func NewRoomProxyAPI(controller string) *RoomProxyAPI {
	return &RoomProxyAPI{
		origin: controller,
	}
}

func (a *RoomProxyAPI) GetRoomMasterProxy(roomID string) (masterProxy model.Proxy, err error) {
	u := url.URL{
		Scheme: "http",
		Host:   a.origin,
		Path:   "/api/v1/room/" + roomID + "/master-proxy",
	}
	_ = u
	panic("todo")
	// res, err := http.Get()
}
func (a *RoomProxyAPI) SetRoomMasterProxy(roomID string, masterProxyID string) (err error) {
	panic("todo")
}
func (a *RoomProxyAPI) GetProxyMasterRooms(proxyID string) (roomID []string, err error) {
	panic("todo")
}
