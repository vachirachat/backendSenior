package message_types

const (
	// Chat has payload of chat message
	Chat = "CHAT"
	// Room has payload indicating room leave or join event
	Room = "ROOM"
	// Error represent error
	Error = "ERROR"

	InvalidateMaster = "INVALIDATE_MASTER"
	InvalidateKey    = "INVALIDATE_KEY"

	// KeyRequest is used for requesting key
	KeyRequest = "KEY_REQUEST"

	// KeyResponse is response for requesting key
	KeyResponse = "KEY_RESPONSE"
)
