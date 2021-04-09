package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type StickerRepository struct {
	conn       *mgo.Session
	stickerSet *mgo.Collection
	sticker    *mgo.Collection
}

// NewStickerRepository is repository for both Sticker and StickerSet
func NewStickerRepository(conn *mgo.Session) *StickerRepository {
	return &StickerRepository{
		conn:       conn,
		sticker:    conn.DB(dbName).C(collectionSticker),
		stickerSet: conn.DB(dbName).C(collectionStickerSet),
	}
}

var (
	_ repository.StickerRepository    = (*StickerRepository)(nil)
	_ repository.StickerSetRepository = (*StickerRepository)(nil)
)

// Implement sticker set repository

func (s StickerRepository) FindStickerSet(filter interface{}) ([]model.StickerSet, error) {
	var res []model.StickerSet
	err := s.stickerSet.Find(filter).All(&res)
	return res, err
}

func (s StickerRepository) GetStickerSetByID(ID bson.ObjectId) (model.StickerSet, error) {
	var res model.StickerSet
	err := s.stickerSet.FindId(ID).One(&res)
	return res, err
}

func (s StickerRepository) InsertStickerSet(stickerSet model.StickerSet) (bson.ObjectId, error) {
	id := bson.NewObjectId()
	stickerSet.ID = id
	if err := s.stickerSet.Insert(stickerSet); err != nil {
		return "", err
	}
	return id, nil
}

// Implement Sticker repository

func (s StickerRepository) FindSticker(filter interface{}) ([]model.Sticker, error) {
	var res []model.Sticker
	err := s.sticker.Find(filter).All(&res)
	return res, err
}

func (s StickerRepository) GetStickerByID(ID bson.ObjectId) (model.Sticker, error) {
	var res model.Sticker
	err := s.sticker.FindId(ID).One(&res)
	return res, err
}

func (s StickerRepository) InsertSticker(sticker model.Sticker) (bson.ObjectId, error) {
	id := bson.NewObjectId()
	sticker.ID = id
	if err := s.sticker.Insert(sticker); err != nil {
		return "", err
	}
	return id, nil
}

func (s StickerRepository) RemoveStickers(filter interface{}) error {
	_, err := s.sticker.RemoveAll(filter)
	return err
}
