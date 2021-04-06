package chat

import (
	"backendSenior/domain/model/chatsocket"
	"backendSenior/domain/model/chatsocket/key_exchange"
	"backendSenior/domain/model/chatsocket/message_types"
	"encoding/json"
	"proxySenior/domain/service"
)

// KeyExhangeController handle websocket message involving key exchange related events
type KeyExhangeController struct {
	ctrl *service.ChatUpstreamService // talking w/ controller
}

// NewKeyExhangeController create new KeyExhangeController, which
//  handle websocket message involving key exchange related events
func NewKeyExhangeController(ctrl *service.ChatUpstreamService) *KeyExhangeController {
	return &KeyExhangeController{
		ctrl: ctrl,
	}
}

func (c *KeyExhangeController) Start() {
	pipe := make(chan []byte, 100)
	c.ctrl.RegisterHandler(pipe)
	defer c.ctrl.UnRegisterHandler(pipe)

	for {
		var msg chatsocket.RawMessage
		data := <-pipe
		err := json.Unmarshal(data, &msg)
		if err != nil {
			//
			continue
		}

		switch msg.Type {
		case message_types.KeyRequest:
			var msg key_exchange.KeyExchangeRequest
			_ = msg
			// case message_types.KeyResponse:

		}
	}
}
