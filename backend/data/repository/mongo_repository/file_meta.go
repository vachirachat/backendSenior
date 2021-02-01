package mongo_repository

type FileMetaRepositoryMongo struct {
	db *mgo.Sesion
}

func NewFileMetaRepositoryMongo(db *mgo.Session) *FileMetaRepositoryMongo {

}

var _ repository.FileMetaRepository = (*FileMetaRepositoryMongo)(nil)

func (r *FileMetaRepositoryMongo) InsertFile(file model.File) error {

}
func (r *FileMetaRepositoryMongo) FindFile(file model.FileFilter) ([]model.File, error) {

}
func (r *FileMetaRepositoryMongo) DeleteFile(file model.FileFilter) error {

}
