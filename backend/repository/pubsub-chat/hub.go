package repository

import (
	"backendSenior/model"
	"backendSenior/utills"
	"log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// NewHub will will give an instance of an Hub

type Hub struct {
	Clients    map[*Client]bool
	Room       map[bson.ObjectId][]*Client
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {

	return &Hub{
		Clients:    make(map[*Client]bool),
		Room:       getAllRoom(),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
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

		}
	}
}

const (
	DBRoomName     = "Room"
	RoomCollection = "RoomData"
)

func getAllRoom() map[bson.ObjectId][]*Client {
	var rooms []model.Room
	var ConnectionDB, err = mgo.Dial(utills.MONGOENDPOINT)
	if err != nil {
		log.Println("error getAllRoom in NewHub()  ", err.Error())
	}
	err = ConnectionDB.DB(DBRoomName).C(RoomCollection).Find(nil).All(&rooms)

	var Map map[bson.ObjectId][]*Client

	for _, room := range rooms {
		for _, userID := range room.ListUser {
			Map[room.RoomID] = append(Map[room.RoomID], &Client{
				userID: userID,
			})
		}
	}
	return Map
}
