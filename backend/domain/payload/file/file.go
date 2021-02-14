package file_payload

type BeforeUploadFileResponse struct {
	URL      string            `json:"url"`
	FileID   string            `json:"fileId"`
	FormData map[string]string `json:"formData"`
}

type BeforeUploadImageResponse struct {
	ImageID  string `json:"imgId"`
	ImageURL string `json:"imgUrl"`
	ThumbID  string `json:"thumbId"`
	ThumbURL string `json:"thumbUrl"`
}
