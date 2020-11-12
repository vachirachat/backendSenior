package socket

import "backendSenior/model"

// NewHub will will give an instance of an Hub
func NewHub() *model.Hub {
	return &model.Hub{
		Register:   make(chan *model.Client),
		Unregister: make(chan *model.Client),
		Clients:    make(map[*model.Client]bool),
	}
}

// Run will execute Go Routines to check incoming Socket events
func Run(hub *model.Hub) {
	for {
		select {
		case client := <-hub.Register:
			HandleUserRegisterEvent(hub, client)

		case client := <-hub.Unregister:
			HandleUserDisconnectEvent(hub, client)
		}
	}
}
