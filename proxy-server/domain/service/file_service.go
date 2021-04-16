package service

import (
	"backendSenior/domain/model"
	file_payload "backendSenior/domain/payload/file"
	"backendSenior/domain/service"
	"common/rmq"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/httputil"
	"net/url"
	"proxySenior/domain/encryption"
	"proxySenior/utils"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/go-resty/resty/v2"
)

type FileService struct {
	host        string
	clnt        *resty.Client
	rabbit      *rmq.RMQClient
	pendingTask map[int]chan bool
}

// NewFileService create new file service, utilizing rabbitMQ
// rabbit MQ must be ready (initailized)
func NewFileService(controllerHost string, rabbit *rmq.RMQClient) *FileService {
	return &FileService{
		host:        controllerHost,
		clnt:        resty.New(),
		rabbit:      rabbit,
		pendingTask: make(map[int]chan bool),
	}
}

func (s *FileService) Run() error {
	msgs, err := s.rabbit.Consume("upload_result")
	if err != nil {
		return fmt.Errorf("consume error: %w", err)
	}

	for {
		m := <-msgs
		var res struct {
			TaskID int
		}
		m.Ack(true)
		err := json.Unmarshal(m.Body, &res)
		if err != nil {
			fmt.Printf("[mq] handle message error: %s\nmessage was: %s\n", err, m.Body)
			continue
		}
		c, ok := s.pendingTask[res.TaskID]
		if !ok {
			fmt.Printf("[mq] WARN: received finish taskID %d but task doesn't exist!", res.TaskID)
			continue
		}
		fmt.Printf("[mq] DONE task %d\n", res.TaskID)
		close(c)
	}

	return nil
}

// GetAnyFile
// fileType: image or file
// fileID: Id of file or thumbnail (it's objectid)
func (s *FileService) GetAnyFile(fileType string, fileID string) (file []byte, err error) {
	u := url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   "/api/v1/file/file-url",
	}
	var res struct {
		URL string `json:"url"`
	}
	err = utils.HTTPPost(u.String(), "application/json", map[string]interface{}{
		"type": fileType,
		"id":   fileID,
	}, &res)
	if err != nil {
		log.Println("get file: get url:", err)
	}

	resp, err := http.Get(res.URL)
	if err != nil {
		return nil, fmt.Errorf("get file: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("get file: non ok: read error: %w", err)
		}
		return nil, fmt.Errorf("get file: request status: %d body: %s", resp.StatusCode, body)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("get file: read error: %w", err)
	}

	return data, nil
}

func (s *FileService) GetAnyFileMeta(fileID string) (meta model.FileMeta, err error) {
	res, err := s.clnt.R().
		SetResult(&meta).
		Get(fmt.Sprintf("http://%s/api/v1/file/by-id/%s", s.host, fileID))
	if !res.IsSuccess() {
		return meta, fmt.Errorf("server response with status: %d\n%s", res.StatusCode(), res.String())
	}
	if err != nil {
		return meta, fmt.Errorf("request failed: %w", err)
	}
	return meta, nil

}

func (s *FileService) ListRoomFiles(roomID string) ([]model.FileMeta, error) {
	u := url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   fmt.Sprintf("/api/v1/file/room/%s/files", roomID),
	}
	var fileMetas []model.FileMeta
	err := utils.HTTPGet(u.String(), &fileMetas)
	return fileMetas, utils.WrapError("list room file: request error: %w", err)
}

func (s *FileService) ListRoomImages(roomID string) ([]model.FileMeta, error) {
	u := url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   fmt.Sprintf("/api/v1/file/room/%s/files", roomID),
	}
	var fileMetas []model.FileMeta
	err := utils.HTTPGet(u.String(), &fileMetas)
	return fileMetas, utils.WrapError("list room file: request error: %w", err)

}

func (s *FileService) BeforeUpload() (file_payload.BeforeUploadFileResponse, error) {
	u := url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   "/api/v1/file/before-upload",
	}

	var res file_payload.BeforeUploadFileResponse
	_, err := s.clnt.R().
		SetResult(&res).
		Post(u.String())
	if err != nil {
		return file_payload.BeforeUploadFileResponse{}, fmt.Errorf("prepare to upload: %w", err)
	}
	return res, nil
}

func (s *FileService) BeforeUploadImage() (file_payload.BeforeUploadImageResponse, error) {
	u := url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   "/api/v1/file/before-upload-image",
	}

	var res file_payload.BeforeUploadImageResponse
	_, err := s.clnt.R().
		SetResult(&res).
		Post(u.String())
	if err != nil {
		return file_payload.BeforeUploadImageResponse{}, fmt.Errorf("prepare to upload: %w", err)
	}
	return res, nil
}

func (s *FileService) waitTaskComplete(taskId int) chan bool {
	s.pendingTask[taskId] = make(chan bool)
	return s.pendingTask[taskId]
}

// FileDetail is detail of file to be send to worker to upload

// UploadFile for handling upload file at specific path
// return file id of new file, <async> fileMeta of uploadFile, error
// error nil if get meta success
// it can still fail
func (s *FileService) UploadFile(roomID string, userID string, key []byte, fileDetail model.FileDetail) (string, chan model.FileMeta, error) {
	// ---- get upload URL
	metaAsync := make(chan model.FileMeta, 1)
	uploadFileRes, err := s.BeforeUpload()
	taskID := rand.Int()

	payload, err := json.Marshal(model.UploadFileTask{
		TaskID:          taskID,
		Type:            model.File,
		FilePath:        fileDetail.Path,
		EncryptKey:      key,
		URL:             uploadFileRes.URL,
		UploadPostForm:  uploadFileRes.FormData,
		UploadPostForm2: nil,
	})
	if err != nil {
		close(metaAsync)
		return "", metaAsync, fmt.Errorf("marshal message to send: %w", err)
	}
	w := s.waitTaskComplete(taskID)

	// send task
	err = s.rabbit.Publish("upload_task", payload)
	if err != nil {
		close(metaAsync)
		return "", metaAsync, fmt.Errorf("publish message: %w", err)
	}

	// wait for task done
	go func() {
		defer close(metaAsync)
		<-w

		newName := addDateToName(fileDetail.Name)
		fileNameEnc, err := encryption.AESEncrypt([]byte(newName), key)
		if err != nil {
			log.Println("upload image: encrypt meta: %w\n", err)
			return
		}
		fileNameEnc = encryption.B64Encode(fileNameEnc)

		meta := model.FileMeta{
			FileID:      bson.ObjectIdHex(uploadFileRes.FileID),
			ThumbnailID: "",
			UserID:      bson.ObjectIdHex(userID),
			RoomID:      bson.ObjectIdHex(roomID),
			BucketName:  "file",
			FileName:    newName,
			Size:        fileDetail.Size,
			CreatedAt:   fileDetail.CreatedTime,
		}
		afterUploadUrl := url.URL{
			Scheme: "http",
			Host:   s.host,
			Path:   fmt.Sprintf("/api/v1/file/room/%s/after-upload/%s", roomID, uploadFileRes.FileID),
		}

		res, err := s.clnt.R().
			SetBody(service.UploadFileMeta{
				Name:      string(fileNameEnc),
				RoomID:    bson.ObjectIdHex(roomID),
				UserID:    bson.ObjectIdHex(userID),
				Size:      fileDetail.Size,
				CreatedAt: fileDetail.CreatedTime,
			}).
			SetHeader("Content-Type", "application/json").
			Post(afterUploadUrl.String())

		if err != nil {
			fmt.Printf("[UPLOAD ERROR] request error: %s", err)
			return
		}
		if !res.IsSuccess() {
			fmt.Printf("[UPLOAD ERROR] after upload: server responded with status %d", res.StatusCode())
			return
			//return metaAsync,
		}

		metaAsync <- meta
	}()

	return uploadFileRes.FileID, metaAsync, nil
}

func addDateToName(fileName string) string {
	parts := strings.Split(fileName, ".")
	idx := len(parts) - 2
	if idx < 0 {
		idx = 0
	}
	parts[idx] = fmt.Sprintf("%s.%d.%d", parts[idx], time.Now().Unix(), rand.Int())
	return strings.Join(parts, ".")
}

func (s *FileService) UploadImage(roomID string, userID string, key []byte, fileDetail model.FileDetail) (string, chan model.FileMeta, error) {
	// ---- get upload URL
	metaAsync := make(chan model.FileMeta, 1)
	uploadImageRes, err := s.BeforeUploadImage()
	taskID := rand.Int()

	payload, err := json.Marshal(model.UploadFileTask{
		TaskID:          taskID,
		Type:            model.Image,
		FilePath:        fileDetail.Path,
		EncryptKey:      key,
		URL:             uploadImageRes.URL,
		UploadPostForm:  uploadImageRes.ImageFormData,
		UploadPostForm2: uploadImageRes.ThumbFormData,
	})
	if err != nil {
		close(metaAsync)
		return "", metaAsync, fmt.Errorf("marshal message to send: %w", err)
	}
	w := s.waitTaskComplete(taskID)

	// send task
	err = s.rabbit.Publish("upload_task", payload)
	if err != nil {
		close(metaAsync)
		return "", metaAsync, fmt.Errorf("publish message: %w", err)
	}

	// wait for task done
	go func() {
		defer close(metaAsync)
		<-w
		newName := addDateToName(fileDetail.Name)
		fileNameEnc, err := encryption.AESEncrypt([]byte(newName), key)
		if err != nil {
			log.Println("upload image: encrypt meta: %w\n", err)
			return
		}
		fileNameEnc = encryption.B64Encode(fileNameEnc)

		// meta is returned to caller (but not sent to backend)
		meta := model.FileMeta{
			FileID:      bson.ObjectIdHex(uploadImageRes.ImageID),
			ThumbnailID: bson.ObjectIdHex(uploadImageRes.ThumbID),
			UserID:      bson.ObjectIdHex(userID),
			RoomID:      bson.ObjectIdHex(roomID),
			BucketName:  "image",
			FileName:    newName,
			Size:        fileDetail.Size,
			CreatedAt:   fileDetail.CreatedTime,
		}
		afterUploadUrl := url.URL{
			Scheme: "http",
			Host:   s.host,
			Path:   fmt.Sprintf("/api/v1/file/room/%s/after-upload-image/%s", roomID, uploadImageRes.ImageID),
		}

		res, err := s.clnt.R().
			SetBody(service.UploadFileMeta{
				Name:      string(fileNameEnc),
				RoomID:    bson.ObjectIdHex(roomID),
				UserID:    bson.ObjectIdHex(userID),
				Size:      fileDetail.Size,
				CreatedAt: fileDetail.CreatedTime,
			}).
			SetHeader("Content-Type", "application/json").
			Post(afterUploadUrl.String())

		if err != nil {
			fmt.Printf("[UPLOAD ERROR] request error: %s", err)
			return
		}
		if !res.IsSuccess() {
			fmt.Printf("[UPLOAD ERROR] after upload: server responded with status %d", res.StatusCode())
			return
			//return metaAsync,
		}

		metaAsync <- meta
	}()

	return uploadImageRes.ImageID, metaAsync, nil
}
