package repository

import (
	"backendSenior/domain/model"
	"github.com/globalsign/mgo/bson"
)

type StickerSetRepository interface {
	FindStickerSet(filter interface{}) ([]model.StickerSet, error)
	GetStickerSetByID(ID bson.ObjectId) (model.StickerSet, error)
	InsertStickerSet(stickerSet model.StickerSet) (bson.ObjectId, error)
	RemoveStickerSets(filter interface{}) (int, error)
	UpdateStickerSetByID(ID bson.ObjectId, update interface{}) error
}

type StickerRepository interface {
	FindSticker(filter interface{}) ([]model.Sticker, error)
	GetStickerByID(ID bson.ObjectId) (model.Sticker, error)
	InsertSticker(sticker model.Sticker) (bson.ObjectId, error)
	RemoveStickers(filter interface{}) (int, error)
	UpdateStickerByID(ID bson.ObjectId, update interface{}) error
	CountSticker(filter interface{}) (int, error)
}
