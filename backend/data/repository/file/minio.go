package file

import (
	"backendSenior/domain/interface/repository"
	"bytes"
	"fmt"
	"github.com/minio/minio-go/pkg/credentials"
	"io"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/minio/minio-go"
)

type MinioStore struct {
	clnt *minio.Client
}

var _ repository.ObjectStore = (*MinioStore)(nil)

const defaultRegion = "us-east-1"

type MinioConfig struct {
	Endpoint  string
	AccessID  string
	SecretKey string
	UseSSL    bool
}

func NewFileStore(config *MinioConfig) (*MinioStore, error) {
	s := &MinioStore{}
	c, err := minio.NewWithOptions(config.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(config.AccessID, config.SecretKey, ""),
		Secure:       false,
		Region:       defaultRegion,
		BucketLookup: 0,
	})
	if err != nil {
		return nil, err
	}
	s.clnt = c
	return s, nil
}

const readWritePolicy = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "AWS": ["*"] },
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads"
      ],
      "Resource": ["arn:aws:s3:::%s"]
    },
    {
      "Effect": "Allow",
      "Principal": { "AWS": ["*"] },
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:DeleteObject",
        "s3:GetObject",
        "s3:ListMultipartUploadParts",
        "s3:PutObject"
      ],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}
`

func (s *MinioStore) ensureBucket(name string) error {

	if exists, err := s.clnt.BucketExists(name); err != nil {
		return err
	} else if !exists {
		err := s.clnt.MakeBucket(name, defaultRegion)
		if err != nil {
			return err
		}
	}

	if err := s.clnt.SetBucketPolicy(name, fmt.Sprintf(readWritePolicy, name, name)); err != nil {
		return err
	}
	return nil

}

func (s *MinioStore) Init() error {
	for _, bucket := range []string{"image", "file", "profile", "sticker", "room"} {
		if err := s.ensureBucket(bucket); err != nil {
			return fmt.Errorf("error ensuring bucket %s: %w", bucket, err)
		}
	}

	return nil
}

const getExpires = time.Second * 60 * 60 * 100
const postExpires = time.Second * 60 * 60 * 100

// GetPresignedURL return URL of getting object
func (s *MinioStore) GetPresignedURL(bucketName string, objectName string) (string, error) {
	url, err := s.clnt.PresignedGetObject(bucketName, objectName, getExpires, url.Values{})
	if err != nil {
		return "", fmt.Errorf("error creating get url", err)
	}
	return url.String(), err
}

// PutPresignedURL return URL for uploading file
func (s *MinioStore) PutPresignedURL(bucketName string, objectName string) (string, error) {
	//url, err := s.clnt.PresignedPutObject(bucketName, objectName, getExpires)
	vals := url.Values{}
	vals.Set("X-Amz-SignedHeaders", "Host")
	vals.Set("Foo", "Bar")
	url, err := s.clnt.Presign("PUT", bucketName, objectName, postExpires, vals)
	if err != nil {
		return "", fmt.Errorf("error creating put url", err)
	}

	return url.String(), err
}

// PostPresignedURL return URL and formData for uploading file using POST
func (s *MinioStore) PostPresignedURL(bucketName string, objectName string) (string, map[string]string, error) {
	policy := minio.NewPostPolicy()

	// Apply upload policy restrictions:
	policy.SetBucket(bucketName)
	policy.SetKey(objectName)
	policy.SetExpires(time.Now().UTC().AddDate(0, 0, 10)) // expires in 10 days

	// Get the POST form key/value object:
	url, formData, err := s.clnt.PresignedPostPolicy(policy)
	if err != nil {
		fmt.Println(err)
		return "", nil, fmt.Errorf("creating presigned post: %w", err)
	}
	return url.String(), formData, nil
}

func (s *MinioStore) DeleteObject(bucketName string, objectName string) (err error) {
	err = s.clnt.RemoveObject(bucketName, objectName)
	return err
}

func (s *MinioStore) PutObject(bucketName string, objectName string, data io.Reader) (err error) {
	b, err := ioutil.ReadAll(data)
	if err != nil {
		return fmt.Errorf("read error: %w", err)
	}
	r := bytes.NewReader(b)
	_, err = s.clnt.PutObject(bucketName, objectName, r, r.Size(), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("upload error: %w", err)
	}
	return nil
}

func (s *MinioStore) GetObject(bucketName string, objectName string) ([]byte, error) {
	obj, err := s.clnt.GetObject(bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	data, err := ioutil.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read object content: %w", err)
	}
	return data, nil
}
