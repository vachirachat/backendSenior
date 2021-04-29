package dto

import (
	"backendSenior/domain/model"

	"github.com/globalsign/mgo/bson"
)

// UpdateMeDto is request body for update user-info(model.USER)
type CreateProxyDto struct {
	IP   string `json:"ip" validate:"required,gt=0"`
	Port int    `json:"port" validate:"required,gt=0"`
	Name string `json:"name"  validate:"required,gt=0"`
	Org  string `json:"org" validate:"required,gt=0"`
}

func (d *CreateProxyDto) ToProxy(secret string) model.Proxy {
	return model.Proxy{
		IP:     d.IP,
		Port:   d.Port,
		Secret: secret,
		Name:   d.Name,
		Rooms:  []bson.ObjectId{},
		Org:    bson.ObjectIdHex(d.Org),
	}
}

// UpdateMeDto is request body for update user-info(model.USER)
type UpdateProxyOrgDto struct {
	Proxies []bson.ObjectId `json:"proxies" validate:"required,gt=0"`
}

func (d *UpdateProxyOrgDto) ToProxyUpdate() model.Organize {
	return model.Organize{
		Proxies: d.Proxies,
	}
}

// UpdateMeDto is request body for update user-info(model.USER)
type UpdateProxyDto struct {
	IP   string `json:"ip" validate:"required,gt=0"`
	Port int    `json:"port" validate:"required,gt=0"`
	Name string `json:"name"  validate:"required,gt=0"`
}

func (d *UpdateProxyDto) ToProxyUpdate() model.Proxy {
	return model.Proxy{
		IP:   d.IP,
		Port: d.Port,
		Name: d.Name,
	}
}
