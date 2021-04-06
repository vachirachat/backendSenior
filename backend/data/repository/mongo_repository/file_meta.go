package mongo_repository

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type FileMetaRepositoryMongo struct {
	conn *mgo.Session
	col  *mgo.Collection
}

func NewFileMetaRepositoryMongo(conn *mgo.Session) *FileMetaRepositoryMongo {
	return &FileMetaRepositoryMongo{
		conn: conn,
		col:  conn.DB(dbName).C(collectionMeta),
	}
}

var _ repository.FileMetaRepository = (*FileMetaRepositoryMongo)(nil)

func (r *FileMetaRepositoryMongo) InsertFile(file model.FileMeta) error {
	err := r.col.Insert(file)
	if err != nil {
		return fmt.Errorf("insert file meta error: %w", err) // TODO: do this for other repo
	}
	return err
}
func (r *FileMetaRepositoryMongo) FindFile(filter model.FileMetaFilter) ([]model.FileMeta, error) {
	var files []model.FileMeta
	err := r.col.Find(filter).All(&files)
	if err != nil {
		return nil, fmt.Errorf("find file meta error: %w", err) // TODO: do this for other repo
	}
	return files, nil
}
func (r *FileMetaRepositoryMongo) DeleteByID(fileID bson.ObjectId) error {
	err := r.col.RemoveId(fileID)
	if err != nil {
		return fmt.Errorf("remove file meta error: %w", err) // TODO: do this for other repo
	}
	return nil
}

func (r *FileMetaRepositoryMongo) DeleteMany(filter model.FileMetaFilter) error {
	_, err := r.col.RemoveAll(filter)
	if err != nil {
		return fmt.Errorf("removemany file meta error: %w", err) // TODO: do this for other repo
	}
	return nil
}
