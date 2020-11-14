package socket

import "github.com/globalsign/mgo/bson"

// NewHub will will give an instance of an Hub

type Hub struct {
	Clients        map[*Client]bool
	Room           map[bson.ObjectId][]*Client
	Register       chan *Client
	Unregister     chan *Client
	RegisterRoom   chan *Client
	UnregisterRoom chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients: make(map[*Client]bool),
		Room:    make(map[bson.ObjectId][]*Client),

		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		RegisterRoom:   make(chan *Client),
		UnregisterRoom: make(chan *Client),
	}
}

// Run will execute Go Routines to check incoming Socket events
func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.Register:
			HandleUserRegisterEvent(hub, client)

		case client := <-hub.Unregister:
			HandleUserDisconnectEvent(hub, client)

		case client := <-hub.RegisterRoom:
			HandleInitConnectRegisterEvent(hub, client)

		}
	}
}
