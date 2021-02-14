package service

import (
	file_payload "backendSenior/domain/payload/file"
	"bytes"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/go-resty/resty/v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/httputil"
	"net/url"
	"proxySenior/utils"
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
func (s *FileService) GetAnyFile(fileType string, fileID string) (file io.Reader, err error) {
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

	return bytes.NewReader(data), nil
}

// func (s *FileService) ListRoomFiles(roomID bson.ObjectId) ([]model.FileMeta, error) {
// 	u := url.URL{
// 		Scheme: "http",
// 		Host:   s.host,
// 		Path:   fmt.Sprintf("/api/v1/file/room/%s/files", roomID.Hex()),
// 	}
// 	var fileMetas []model.FileMeta
// 	err := utils.HTTPGet(u.String(), &fileMetas)
// 	return fileMetas, utils.WrapError("list room file: request error: %w", err)
// }

// func (s *FileService) ListRoomImages(roomID bson.ObjectId) ([]model.FileMeta, error) {
// 	u := url.URL{
// 		Scheme: "http",
// 		Host:   s.host,
// 		Path:   fmt.Sprintf("/api/v1/file/room/%s/files", roomID.Hex()),
// 	}
// 	var fileMetas []model.FileMeta
// 	err := utils.HTTPGet(u.String(), &fileMetas)
// 	return fileMetas, utils.WrapError("list room file: request error: %w", err)

// }

func (s *FileService) UploadFile(roomID bson.ObjectId, userID bson.ObjectId, filename string, file io.ReadCloser) error {
	// ---- get upload URL
	u := url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   "/api/v1/file/before-upload",
	}
	var uploadFileRes file_payload.BeforeUploadFileResponse
	err := utils.HTTPPost(u.String(), "application/json", "", &uploadFileRes)
	if err != nil {
		return fmt.Errorf("upload file: get url: %w", err)
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
		return err
	}
	fmt.Println("res:", res.String())

	//if err != nil {
	//	return fmt.Errorf("upload file: upload to minio: %w", err)
	//}
	//fmt.Println(uploadRes.Request.Header)
	//defer uploadRes.Body.Close()
	//
	//if uploadRes.StatusCode >= 400 {
	//	body, err := ioutil.ReadAll(uploadRes.Body)
	//	if err != nil {
	//		return fmt.Errorf("upload file: upload to minio: non ok %d: read error: %w", uploadRes.StatusCode, err)
	//	}
	//	return fmt.Errorf("upload file: upload to minio: request status: %d body: %s", uploadRes.StatusCode, body)
	//}
	//
	//afterUploadUrl := url.URL{
	//	Scheme: "http",
	//	Host:   s.host,
	//	Path:   fmt.Sprintf("/api/v1/file/room/%s/after-upload/%s", roomID.Hex(), uploadFileRes.FileID),
	//}
	//
	//var res interface{}
	//err = utils.HTTPPost(afterUploadUrl.String(), "application/json", map[string]interface{}{ // TODO: refactor
	//	"name":   filename, // name of file
	//	"roomId": roomID,   // room to associate file
	//	"userId": userID,   // file owner
	//	"size":   -1,
	//}, &res)


	return utils.WrapError("upload file: after upload: %w", err)
}

func (s *FileService) UploadImage() {
	// TODO: this require making thumbnail
}
