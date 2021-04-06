package repository

import (
	"backendSenior/domain/model"

	"github.com/globalsign/mgo/bson"
)

//  FileMetaRepository repository for storing fiel
type FileMetaRepository interface {
	InsertFile(file model.FileMeta) error
	FindFile(file model.FileMetaFilter) ([]model.FileMeta, error)
	DeleteByID(fileID bson.ObjectId) error
	DeleteMany(filter model.FileMetaFilter) error
}
