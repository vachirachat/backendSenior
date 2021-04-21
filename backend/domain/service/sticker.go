package service

import (
	"backendSenior/domain/dto"
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"bytes"
	"common/utils/db"
	"errors"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"log"
	"os"
)

type StickerService struct {
	stickerRepo    repository.StickerRepository
	stickerSetRepo repository.StickerSetRepository
	objectStore    repository.ObjectStore
	l              *log.Logger
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
		l:              log.New(os.Stdout, "StickerService", log.LstdFlags|log.Lshortfile),
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

func (s *StickerService) AddStickerToSet(setID bson.ObjectId, meta dto.CreateStickerDto, imgData []byte) (bson.ObjectId, error) {
	_, err := s.stickerSetRepo.GetStickerSetByID(setID)
	if err != nil {
		return "", err
	}

	id, err := s.stickerRepo.InsertSticker(model.Sticker{
		SetID: setID,
		Name:  meta.Name,
	})
	if err != nil {
		return "", fmt.Errorf("add to database: %w", err)
	}

	if err := s.objectStore.PutObject("sticker", id.Hex(), bytes.NewReader(imgData)); err != nil {
		return "", err
	}
	return id, nil
}

func (s *StickerService) GetStickerImage(ID bson.ObjectId) ([]byte, error) {
	if exists, err := s.StickerExists(ID); err != nil {
		return nil, fmt.Errorf("can't check sticker: %w", err)
	} else if !exists {
		return nil, errors.New("sticker not exist")
	}

	data, err := s.objectStore.GetObject("sticker", ID.Hex())
	if err != nil {
		return nil, fmt.Errorf("reading sticker image: %w", err)
	}
	return data, nil
}

func (s *StickerService) RemoveSticker(ID bson.ObjectId) error {
	if cnt, err := s.stickerRepo.RemoveStickers(model.StickerFilter{ID: ID}); err != nil {
		s.l.Println("[ERROR] error deleting sticker meta", err)
		return fmt.Errorf("error deleting sticker: %w", err)
	} else if cnt == 0 {
		return fmt.Errorf("sticker not found")
	}

	if err := s.objectStore.DeleteObject("sticker", ID.Hex()); err != nil {
		s.l.Println("[WARN] error deleting sticker file, file will be \"dangling\":", err)
	}

	return nil
}

func (s *StickerService) RemoveStickerSet(ID bson.ObjectId) error {
	if cnt, err := s.stickerSetRepo.RemoveStickerSets(model.StickerSetFilter{ID: ID}); err != nil {
		s.l.Println("[ERROR] error deleting sticker set meta", err)
		return fmt.Errorf("error deleting sticker set: %w", err)
	} else if cnt == 0 {
		return fmt.Errorf("sticker not found")
	}
	return nil
}

func (s *StickerService) EditStickerSetInfo(ID bson.ObjectId, editInfo dto.EditStickerSetDto) error {
	err := s.stickerSetRepo.UpdateStickerSetByID(ID, model.StickerSetFilter{
		Name: editInfo.Name,
	})
	return err
}

func (s *StickerService) EditStickerInfo(ID bson.ObjectId, editInfo dto.EditStickerDto) error {
	err := s.stickerRepo.UpdateStickerByID(ID, model.StickerFilter{
		Name: editInfo.Name,
	})
	return err
}

func (s *StickerService) StickerSetIsEmpty(setID bson.ObjectId) (bool, error) {
	cnt, err := s.stickerRepo.CountSticker(model.StickerFilter{
		SetID: setID,
	})
	if err != nil {
		return false, err
	}
	return cnt == 0, nil
}

func (s *StickerService) FindStickerSet(query dto.FindStickerSetDto) ([]model.StickerSet, error) {
	return s.stickerSetRepo.FindStickerSet(model.StickerSetFilter{
		Name: db.Contains(query.Name, db.CaseInsensitive),
	})
}
