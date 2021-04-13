package dto

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type UploadFileMeta struct {
	Name      string        `json:"fileName" validate:"required,gt=0"` // name of file
	RoomID    bson.ObjectId // room to associate file
	UserID    bson.ObjectId // file owner
	Size      int           `json:"size" validate:"required,gt=0"`
	CreatedAt time.Time     `json:"createdAt" validate:"required,gt=0"` // time that file is encrypted at proxy
}

type UploadImageMeta struct {
	Name        string        `json:"fileName" validate:"required,gt=0"` // name of file
	RoomID      bson.ObjectId // room to associate file
	UserID      bson.ObjectId // file owner
	Size        int           `json:"size" validate:"required,gt=0"`
	ThumbnailID bson.ObjectId
}

type FileQuery struct {
	Type   string        `json:"type" validate:"required,gt=0"`
	FileID bson.ObjectId `json:"id"`
}
