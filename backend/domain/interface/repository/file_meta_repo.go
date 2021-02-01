package repository

//  FileMetaRepository repository for storing fiel
type FileMetaRepository interface {
	InsertFile(file model.File) error
	FindFile(file model.FileFilter) ([]model.File, error)
	DeleteFile(file model.FileFilter) error
}
