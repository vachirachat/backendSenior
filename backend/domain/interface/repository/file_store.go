package repository

import "io"

type ObjectStore interface {
	PutObject(bucketName string, objectName string, file io.Reader) (size int, err error)
	GetObject(bucketName string, objectName string) (file io.Reader, err error)
}
