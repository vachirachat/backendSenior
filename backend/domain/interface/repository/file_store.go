package repository

import "io"

type ObjectStore interface {
	GetPresignedURL(bucketName string, objectName string) (url string, err error)
	PutPresignedURL(bucketName string, objectName string) (url string, err error)
	PostPresignedURL(bucketName string, objectName string) (url string, formData map[string]string, err error)
	DeleteObject(bucketName string, objectName string) (err error)
	PutObject(bucketName string, objectName string, data io.Reader) (err error)
	GetObject(bucketName string, objectName string) ([]byte, error)
	//GetPresignedURL(bucketName string, objectName string) (size int, err error)
	//GetObject(bucketName string, objectName string) (file io.Reader, err error)
}
