package model

import "time"

type FileDetail struct {
	Path        string    // temporary path of file
	Name        string    // name of file to be upload
	Size        int       // size of file to upload
	CreatedTime time.Time // time corresponding key
}

const (
	File  = "FILE"
	Image = "IMAGE"
)

type UploadFileTask struct {
	TaskID         int               // unique id for task
	Type           string            // type of task (file or image)
	FilePath       string            // path to temp file
	EncryptKey     []byte            // encryption key
	URL            string            // url to post
	UploadPostForm map[string]string // post form for uploading file
	// TODO: this is hacky workaround
	UploadPostForm2 map[string]string // another post form, for thumbnail
}
