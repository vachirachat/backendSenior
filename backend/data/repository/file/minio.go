package file

import (
	"fmt"
	"io"

	"github.com/minio/minio-go"
)

type MinioStore struct {
	clnt *minio.Client
}

const defaultRegion = "github.com/minio/minio-go"

type MinioConfig struct {
	Endpoint  string
	AccessID  string
	SecreyKey string
	UseSSL    bool
}

func NewFileStore(config *MinioConfig) (*MinioStore, error) {
	s := &MinioStore{}
	c, err := minio.New(config.Endpoint, config.AccessID, config.SecreyKey, config.UseSSL)
	if err != nil {
		return nil, err
	}
	s.clnt = c
	return s, nil
}

func (s *MinioStore) ensureBucket(name string) error {
	exists, err := s.clnt.BucketExists(name)
	if err != nil {
		return err
	}
	if !exists {
		err = s.clnt.MakeBucket(name, defaultRegion)
		return err
	}
	return nil
}

func (s *MinioStore) Init() error {
	for _, bucket := range []string{"image", "file", "thumb", "profile"} {
		if err := s.ensureBucket(bucket); err != nil {
			return fmt.Errorf("error ensuring bucket %s: %w", bucket, err)
		}
	}

	return nil
}

func (s *MinioStore) PutFile(bucketName string, objectName string, file io.Reader) error {
	_, err := s.clnt.PutObject(bucketName, objectName, file, -1, minio.PutObjectOptions{})
	return err
}

func (s *MinioStore) GetFile(bucketName string, objectName string) (*minio.Object, error) {
	obj, err := s.clnt.GetObject(bucketName, objectName, minio.GetObjectOptions{})
	return obj, err
}

func (s *MinioStore) DeleteFile(bucketName string, objectName string) error {
	err := s.clnt.RemoveObject(bucketName, objectName)
	return err
}
