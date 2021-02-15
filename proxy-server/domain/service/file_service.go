package service

import (
	"backendSenior/domain/model"
	file_payload "backendSenior/domain/payload/file"
	"bytes"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/httputil"
	"net/url"
	"proxySenior/utils"
	"time"
)

type FileService struct {
	host string
	clnt *resty.Client
}

func NewFileService(controllerHost string) *FileService {
	return &FileService{
		host: controllerHost,
		clnt: resty.New(),
	}
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

func (s *FileService) UploadFile(roomID string, userID string, filename string, timestamp time.Time, file io.Reader) (fileID string, err error) {
	// ---- get upload URL
	u := url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   "/api/v1/file/before-upload",
	}
	var uploadFileRes file_payload.BeforeUploadFileResponse
	err = utils.HTTPPost(u.String(), "application/json", "", &uploadFileRes)
	if err != nil {
		return "", fmt.Errorf("upload file: get url: %w", err)
	}

	allFileData, _ := ioutil.ReadAll(file)

	req := s.clnt.R().
		SetFormData(uploadFileRes.FormData).
		SetFileReader("file", "foo.txt", bytes.NewReader(allFileData)).
		SetHeader("Content-Type", "multipart/form-data")
	//
	//req.Header["host"] = []string{"localhost:9000"}

	res, err := req.Post(uploadFileRes.URL)

	if err != nil {
		fmt.Println("resty req error:", err)
		return "", err
	}
	if !res.IsSuccess() {
		return "", fmt.Errorf("response status: %d\n%s", res.StatusCode(), res.String())
	}

	afterUploadUrl := url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   fmt.Sprintf("/api/v1/file/room/%s/after-upload/%s", roomID, uploadFileRes.FileID),
	}

	res, err = s.clnt.R().
		SetBody(map[string]interface{}{ // TODO: refactor
			"name":      filename, // name of file
			"roomId":    roomID,   // room to associate file
			"userId":    userID,   // file owner
			"size":      len(allFileData),
			"createdAt": timestamp,
		}).
		SetHeader("Content-Type", "application/json").
		Post(afterUploadUrl.String())

	if !res.IsSuccess() {
		return "", fmt.Errorf("response status: %d\n%s", res.StatusCode(), res.String())
	}

	return uploadFileRes.FileID, utils.WrapError("upload file: after upload: %w", err)
}

func (s *FileService) UploadImage() {
	// TODO: this require making thumbnail
}
