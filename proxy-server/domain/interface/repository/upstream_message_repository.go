package repository

// UpstreamMessageRepository is interface to send message to upstream (controller)
type UpstreamMessageRepository interface {
	SendMessage(data []byte) error
	RegisterHandler(channel chan []byte) error
	UnRegisterHandler(channel chan []byte) error
	// OnConnect register channel to be notified when upstream is connected
	OnConnect(channel chan struct{})
	// OffConnect unregister channel from being notified when upstream is connected
	OffConnect(channel chan struct{})
	// OnDisconnect register channel to be notified when upstream is disconnected
	OnDisconnect(channel chan struct{})
	// OffDisconnect unregister channel from being notified when upstream is disconnected
	OffDisconnect(channel chan struct{})
}
