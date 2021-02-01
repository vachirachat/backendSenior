package model

type FileMeta struct {
	FileID     bson.objectId `json:"id" bson:"_id"`
	RoomID     bson.objectId `json:"roomId" bson:"roomId"`
	BucketName string        `json:"bucketName" bson:"bucketName"`
	// meta
	FileName  string    `json:"fileName" bson:"fileName"`
	Size      int       `json:"size" bson:"size"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

type FileMetaFilter struct {
	FileID     bson.objectId `bson:"_id,omitempty"`
	RoomID     bson.objectId `bson:"roomId,omitempty"`
	BucketName string        `bson:"bucketName,omitempty"`
	// meta
	FileName  string    `bson:"fileName,omitempty"`
	Size      int       `bson:"size,omitempty"`
	CreatedAt time.Time `bson:"createdAt,omitempty"`
}
