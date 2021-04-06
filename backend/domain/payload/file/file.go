package file_payload

type BeforeUploadFileResponse struct {
	URL      string            `json:"url"`
	FileID   string            `json:"fileId"`
	FormData map[string]string `json:"formData"`
}

type BeforeUploadImageResponse struct {
	URL           string
	ImageID       string            `json:"imgId"`
	ImageFormData map[string]string `json:"imgForm"`
	ThumbID       string            `json:"thumbId"`
	ThumbFormData map[string]string `json:"thumbForm"`
}
