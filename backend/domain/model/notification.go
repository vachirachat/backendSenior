package model

// Notification defines detail for notification sent to user device
type Notification struct {
	Title    string
	Body     string
	ImageURL string
	Data     map[string]string
}
