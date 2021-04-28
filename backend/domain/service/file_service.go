package service

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	file_payload "backendSenior/domain/payload/file"
	"bytes"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/globalsign/mgo"
	"image"
	"io"
	"log"
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
	Name      string        // name of file
	RoomID    bson.ObjectId // room to associate file
	UserID    bson.ObjectId // file owner
	Size      int
	CreatedAt time.Time // time that file is encrypted at proxy
}

type UploadImageMeta struct {
	Name        string        // name of file
	RoomID      bson.ObjectId // room to associate file
	UserID      bson.ObjectId // file owner
	Size        int
	CreatedAt   time.Time // time that file is encrypted at proxy
	ThumbnailID bson.ObjectId
}

func (s *FileService) BeforeUploadFilePOST() (fileID string, endpoint string, formData map[string]string, err error) {
	oid := bson.NewObjectId().Hex()
	endpoint, formData, err = s.file.PostPresignedURL("file", oid)
	if err != nil {
		return "", "", nil, err
	}
	return oid, endpoint, formData, err
}

// AfterUploadFile should be used after uploading file to set meta data to database
func (s *FileService) AfterUploadFile(fileID string, meta UploadFileMeta) error {
	err := s.meta.InsertFile(model.FileMeta{
		FileID:     bson.ObjectIdHex(fileID),
		RoomID:     meta.RoomID,
		BucketName: "file",
		FileName:   meta.Name,
		Size:       meta.Size,
		CreatedAt:  meta.CreatedAt,
		UserID:     meta.UserID,
	})
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}

	return nil
}

func (s *FileService) GetAnyFileMeta(fileID bson.ObjectId) (model.FileMeta, error) {
	metas, err := s.meta.FindFile(model.FileMetaFilter{
		FileID: fileID,
	})
	if err != nil {
		return model.FileMeta{}, fmt.Errorf("error getting file meta: %w", err)
	}
	if len(metas) == 0 {
		return model.FileMeta{}, mgo.ErrNotFound
	}
	// TODO: assume if not error then there's result, check to see if this is true
	return metas[0], nil
}

func (s *FileService) fileExists(fileID bson.ObjectId) (bool, error) {
	// TODO: optimze
	metas, err := s.meta.FindFile(model.FileMetaFilter{FileID: fileID})
	if err != nil {
		return false, err
	}
	return len(metas) > 0, nil
}

// GetFile return file by ID, only valid for type = file
// Don't forget to close file
func (s *FileService) GetFileURL(fileID bson.ObjectId) (string, error) {
	if ok, err := s.fileExists(fileID); !ok || err != nil {
		return "", err
	}
	url, err := s.file.GetPresignedURL("file", fileID.Hex())
	return url, err
}

// GetFile return file by ID, only valid for type = file
func (s *FileService) GetRoomFileMetas(roomID bson.ObjectId) ([]model.FileMeta, error) {
	metas, err := s.meta.FindFile(model.FileMetaFilter{
		RoomID:     roomID,
		BucketName: "file",
	})
	return metas, err
}

// ==================== IMAGE

func (s *FileService) BeforeUploadImagePOST() (file_payload.BeforeUploadImageResponse, error) {
	res := file_payload.BeforeUploadImageResponse{}

	oid := bson.NewObjectId().Hex()
	endpoint, formData, err := s.file.PostPresignedURL("image", oid)
	res.URL = endpoint
	res.ImageFormData = formData
	res.ImageID = oid
	if err != nil {
		return file_payload.BeforeUploadImageResponse{}, fmt.Errorf("generate presigned post: %w", err)
	}

	oid = bson.NewObjectId().Hex()
	endpoint, formData, err = s.file.PostPresignedURL("image", oid)
	res.ThumbFormData = formData
	res.ThumbID = oid
	if err != nil {
		return file_payload.BeforeUploadImageResponse{}, fmt.Errorf("generate presigned post thumbnail: %w", err)
	}

	return res, nil
}

// AfterUploadImage should be used after uploading file to set meta data to database
func (s *FileService) AfterUploadImage(imageFileID string, meta UploadImageMeta) error {
	now := time.Now()

	err := s.meta.InsertFile(model.FileMeta{
		FileID:      bson.ObjectIdHex(imageFileID),
		RoomID:      meta.RoomID,
		BucketName:  "image",
		FileName:    meta.Name,
		Size:        meta.Size,
		CreatedAt:   now,
		UserID:      meta.UserID,
		ThumbnailID: meta.ThumbnailID,
	})
	if err != nil {
		return fmt.Errorf("error uploading image: %w", err)
	}

	return nil
}

func (s *FileService) GetAnyFileURL(fileType string, fileID string) (string, error) {
	url, err := s.file.GetPresignedURL(fileType, fileID)
	if err != nil {
		return url, fmt.Errorf("error getting url: %w", err)
	}
	return url, nil
}

// GetImageURLs return URL for image and it's thumbnail
func (s *FileService) GetImageURLs(fileID bson.ObjectId) (img, thumb string, err error) {
	metas, err := s.meta.FindFile(model.FileMetaFilter{FileID: fileID})
	if err != nil {
		return "", "", fmt.Errorf("error checking image file exists: %w", err)
	}
	if len(metas) == 0 {
		return "", "", fmt.Errorf("image file doesn't exist")
	}

	meta := metas[0]

	img, err = s.file.GetPresignedURL("image", meta.FileID.Hex())
	thumb, err = s.file.GetPresignedURL("image", meta.ThumbnailID.Hex())
	return img, thumb, err
}

// GetRoomImageMetas return meta of images in the room
func (s *FileService) GetRoomImageMetas(roomID bson.ObjectId) ([]model.FileMeta, error) {
	metas, err := s.meta.FindFile(model.FileMetaFilter{
		RoomID:     roomID,
		BucketName: "image",
	})
	return metas, err
}

func (s *FileService) DeleteFile(fileID bson.ObjectId) error {
	if metas, err := s.meta.FindFile(model.FileMetaFilter{
		FileID:     fileID,
		BucketName: "file",
	}); err != nil {
		return fmt.Errorf("finding meta: %w", err)
	} else if len(metas) == 0 {
		return mgo.ErrNotFound
	} else {
		m := metas[0]
		if err := s.file.DeleteObject(m.BucketName, m.FileID.Hex()); err != nil {
			return fmt.Errorf("deleting file: %w", err)
		}
		if err := s.meta.DeleteByID(fileID); err != nil {
			return fmt.Errorf("deleting file meta: %w", err)
		}
		return nil
	}
}

func (s *FileService) DeleteImage(imageFileID bson.ObjectId) error {
	if metas, err := s.meta.FindFile(model.FileMetaFilter{
		FileID:     imageFileID,
		BucketName: "image",
	}); err != nil {
		return fmt.Errorf("finding meta: %w", err)
	} else if len(metas) == 0 {
		return errors.New("file not found")
	} else {
		m := metas[0]
		if err := s.file.DeleteObject(m.BucketName, m.FileID.Hex()); err != nil {
			return fmt.Errorf("deleting main image: %w", err)
		}
		if err := s.file.DeleteObject(m.BucketName, m.ThumbnailID.Hex()); err != nil {
			return fmt.Errorf("deleting thumbnail image: %w", err)
		}
		if err := s.meta.DeleteByID(imageFileID); err != nil {
			return fmt.Errorf("deleting file meta: %w", err)
		}
		return nil
	}
}

func (s *FileService) UploadProfileImage(userID string, file []byte) error {

	r := bytes.NewReader(file)

	// decode & process
	r.Seek(0, io.SeekStart) // reset seek
	src, err := imaging.Decode(r)
	if err != nil {
		log.Printf("imaging: error decoing image: %s, is it corrupt?", err)
		return fmt.Errorf("imaging: error decoing image: %s, is it corrupt?", err)
	}

	size := src.Bounds().Size()
	width := size.X
	height := size.Y
	var img *image.NRGBA
	if width > height {
		img = imaging.Resize(src, 400, 0, imaging.Lanczos)
	} else {
		img = imaging.Resize(src, 0, 400, imaging.Lanczos)
	}

	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, img, imaging.JPEG) // force JPEG !!
	if err != nil {
		log.Printf("error encoding to %s:%s\n", imaging.JPEG, err)
		return fmt.Errorf("error encoding to %s:%s", imaging.JPEG, err)
	}

	r.Seek(0, io.SeekStart)
	if err := s.file.PutObject("profile", "profile_"+userID, r); err != nil {
		return fmt.Errorf("error uploading full image, %w", err)
	}
	if err := s.file.PutObject("profile", "thumb_"+userID, buf); err != nil {
		return fmt.Errorf("error uploading thumbnail image, %w", err)
	}

	return nil
}

func (s *FileService) GetProfileImage(userID string, isThumbnail bool) ([]byte, error) {
	bucket := "profile"
	var fileName string
	if isThumbnail {
		fileName = "thumb_" + userID
	} else {
		fileName = "profile_" + userID
	}

	data, err := s.file.GetObject(bucket, fileName)
	if err != nil {
		return nil, fmt.Errorf("error getting file: %w", err)
	}
	return data, nil

}

func (s *FileService) UploadRoomImage(roomID string, file []byte) error {

	r := bytes.NewReader(file)

	// decode & process
	r.Seek(0, io.SeekStart) // reset seek
	src, err := imaging.Decode(r)
	if err != nil {
		log.Printf("imaging: error decoing image: %s, is it corrupt?", err)
		return fmt.Errorf("imaging: error decoing image: %s, is it corrupt?", err)
	}

	size := src.Bounds().Size()
	width := size.X
	height := size.Y
	var img *image.NRGBA
	if width > height {
		img = imaging.Resize(src, 400, 0, imaging.Lanczos)
	} else {
		img = imaging.Resize(src, 0, 400, imaging.Lanczos)
	}

	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, img, imaging.JPEG) // force JPEG !!
	if err != nil {
		log.Printf("error encoding to %s:%s\n", imaging.JPEG, err)
		return fmt.Errorf("error encoding to %s:%s", imaging.JPEG, err)
	}

	r.Seek(0, io.SeekStart)
	if err := s.file.PutObject("room", "room_"+roomID, r); err != nil {
		return fmt.Errorf("error uploading full image, %w", err)
	}
	if err := s.file.PutObject("room", "thumb_"+roomID, buf); err != nil {
		return fmt.Errorf("error uploading thumbnail image, %w", err)
	}

	return nil
}

func (s *FileService) GetRoomImage(roomID string, isThumbnail bool) ([]byte, error) {
	bucket := "room"
	var fileName string
	if isThumbnail {
		fileName = "thumb_" + roomID
	} else {
		fileName = "room_" + roomID
	}

	data, err := s.file.GetObject(bucket, fileName)
	if err != nil {
		return nil, fmt.Errorf("error getting file: %w", err)
	}
	return data, nil

}
