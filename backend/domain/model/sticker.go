package model

import "github.com/globalsign/mgo/bson"

// StickerSet a set of stickers
type StickerSet struct {
	ID   bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name string        `json:"name" bson:"name"`
}

// Sticker represent a sticker in a set
type Sticker struct {
	ID    bson.ObjectId `json:"id" bson:"_id,omitempty"`
	SetID bson.ObjectId `json:"setId" bson:"setId"`
	Name  string        `json:"name" bson:"name"`
}

type StickerSetFilter struct {
	ID   interface{} `bson:"_id,omitempty"`
	Name interface{} `bson:"name,omitempty"`
}

type StickerFilter struct {
	ID    interface{} `bson:"_id,omitempty"`
	SetID interface{} `bson:"setId,omitempty"`
	Name  interface{} `bson:"name,omitempty"`
}
