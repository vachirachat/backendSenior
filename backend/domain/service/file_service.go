package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"fmt"
	"io"
	"time"

	"github.com/globalsign/mgo/bson"
)

type FileService struct {
	file repository.ObjectStore
	meta repository.FileMetaRepository
}

func NewFileService(file repository.ObjectStore, meta repository.FileMetaRepository) *FileService {
	return &FileService{
		file: file,
		meta: meta,
	}
}

type UploadFileMeta struct {
	Name   string        // name of file
	RoomID bson.ObjectId // room to associate file
}

func (s *FileService) UploadFile(file io.Reader, meta UploadFileMeta) error {
	oid := bson.NewObjectId()
	now := time.Now()
	size, err := s.file.PutObject("file", oid.Hex(), file)
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}

	err = s.meta.InsertFile(model.FileMeta{
		FileID:     oid,
		RoomID:     meta.RoomID,
		BucketName: "file",
		FileName:   meta.Name,
		Size:       size,
		CreatedAt:  now,
	})
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}

	return nil
}

func (s *FileService) GetFileMeta(fileID bson.ObjectId) (model.FileMeta, error) {
	meta, err := s.meta.FindFile(model.FileMetaFilter{
		FileID: fileID,
	})
	if err != nil {
		return meta, fmt.Errorf("error getting file meta: %w", err)
	}
	return meta, nil
}
