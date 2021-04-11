package dto

import "github.com/globalsign/mgo/bson"

// CreateStickerSetDto is request body for creating sticker set
type CreateStickerSetDto struct {
	Name string `json:"name" validate:"required"`
}

// RemoveStickersDto is request body for delete sticker
type RemoveStickersDto struct {
	IDs []bson.ObjectId `json:"ids" validate:"required,min=1,dive,required"`
}

// CreateStickerDto is meta data for creating sticker
// TODO: might be more field in sticker in the future
type CreateStickerDto struct {
	Name string `json:"name" validate:"required"`
}

// EditStickerDto for editing sticker
type EditStickerDto struct {
	Name string `json:"name" validate:"required"`
}

// EditStickerSetDto for editing sticker Set
type EditStickerSetDto struct {
	Name string `json:"name" validate:"required"`
}

type FindStickerSetDto struct {
	Name string `json:"name" validate:"required,min=3"`
}
