package service

import (
	"backendSenior/domain/dto"
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"bytes"
	"errors"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type StickerService struct {
	stickerRepo    repository.StickerRepository
	stickerSetRepo repository.StickerSetRepository
	objectStore    repository.ObjectStore
}

func NewStickerService(
	stickerRepo repository.StickerRepository,
	stickerSetRepo repository.StickerSetRepository,
	objectStore repository.ObjectStore,
) *StickerService {
	return &StickerService{
		stickerRepo:    stickerRepo,
		stickerSetRepo: stickerSetRepo,
		objectStore:    objectStore,
	}
}

func (s *StickerService) NewStickerSet(dto dto.CreateStickerSetDto) (bson.ObjectId, error) {
	sticker := model.StickerSet{
		Name: dto.Name,
	}
	id, err := s.stickerSetRepo.InsertStickerSet(sticker)
	return id, err
}

func (s *StickerService) ListStickerSet() ([]model.StickerSet, error) {
	stickerSets, err := s.stickerSetRepo.FindStickerSet(nil)
	return stickerSets, err
}

func (s *StickerService) GetStickerInSet(setID bson.ObjectId) ([]model.Sticker, error) {
	stickers, err := s.stickerRepo.FindSticker(model.StickerFilter{SetID: setID})
	return stickers, err
}

func (s *StickerService) StickerExists(ID bson.ObjectId) (bool, error) {
	if _, err := s.stickerRepo.GetStickerByID(ID); err == nil {
		return true, nil
	} else if errors.Is(err, mgo.ErrNotFound) {
		return false, nil
	} else {
		return false, err
	}
}

func (s *StickerService) AddStickerToSet(setID bson.ObjectId, meta dto.CreateStickerDto, image []byte) (bson.ObjectId, error) {
	id, err := s.stickerRepo.InsertSticker(model.Sticker{
		SetID: setID,
		// TODO: use data from meta later
	})
	if err != nil {
		return "", fmt.Errorf("add to database: %w", err)
	}
	reader := bytes.NewReader(image)
	if err := s.objectStore.PutObject("sticker", fmt.Sprintf("%s/%s", setID, id), reader); err != nil {
		return "", err
	}
	return id, nil
}
