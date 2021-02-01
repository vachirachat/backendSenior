package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type FileMeta struct {
	FileID     bson.ObjectId `json:"id" bson:"_id"`
	RoomID     bson.ObjectId `json:"roomId" bson:"roomId"`
	BucketName string        `json:"bucketName" bson:"bucketName"`
	// meta
	FileName  string    `json:"fileName" bson:"fileName"`
	Size      int       `json:"size" bson:"size"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

type FileMetaFilter struct {
	FileID     interface{} `bson:"_id,omitempty"`
	RoomID     interface{} `bson:"roomId,omitempty"`
	BucketName interface{} `bson:"bucketName,omitempty"`
	// meta
	FileName  interface{} `bson:"fileName,omitempty"`
	Size      interface{} `bson:"size,omitempty"`
	CreatedAt interface{} `bson:"createdAt,omitempty"`
}
