package backup

// BackupService defines functions available over GRPC
type BackupService interface {
	OnMessageIn(message RawMessage) error
	IsReady() (bool, error)
}

// RawMessage is message received over GRPC
type RawMessage struct {
	MessageID string
	TimeStamp int64
	RoomID    string
	UserID    string
	ClientUID string
	Data      string
	Type      string
}
