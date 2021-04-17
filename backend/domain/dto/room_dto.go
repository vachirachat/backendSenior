package dto

import (
	"backendSenior/domain/model"
	"time"

	"github.com/globalsign/mgo/bson"
)

type CreateGroupDto struct {
	RoomName string        `validate:"required"`
	OrgId    bson.ObjectId `validate:"required"`
}

func (d *CreateGroupDto) ToRoom() model.Room {
	return model.Room{
		RoomName: d.RoomName,
		// We dont have orgId here, since we want it to be set after room is "invited" to org
		CreatedTimeStamp: time.Now(),
		RoomType:         model.RoomGroup,
		ListUser:         []bson.ObjectId{},
		ListProxy:        []bson.ObjectId{},
	}
}

type CreatePrivateChatDto struct {
	RoomName string        `validate:"required"`
	OrgId    bson.ObjectId `validate:"required"`
	UserId   bson.ObjectId `validate:"required"`
}

func (d *CreatePrivateChatDto) ToRoom() model.Room {
	return model.Room{
		RoomName:         d.RoomName,
		OrgID:            d.OrgId,
		CreatedTimeStamp: time.Now(),
		RoomType:         model.RoomPrivate,
		ListUser:         nil,
		ListProxy:        nil,
	}
}

type InviteAdminDto struct {
	UserIDs []bson.ObjectId `validate:"required,min=1,dive,required"`
}

type Empty struct{}
type Any = interface{}
