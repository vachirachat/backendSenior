package repository

// UpstreamMessageRepository is interface to send message to upstream (controller)
type UpstreamMessageRepository interface {
	SendMessage(data []byte) error
	RegisterHandler(channel chan []byte) error
	UnRegisterHandler(channel chan []byte) error
}
