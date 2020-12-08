package repository

// import (
// 	"backendSenior/domain/model"
// 	"backendSenior/utills"
// 	"log"

// 	"github.com/globalsign/mgo"
// 	"github.com/globalsign/mgo/bson"
// )

// // NewHub will will give an instance of an Hub

// type Hub struct {
// 	Clients    map[*Client]bool
// 	Room       map[bson.ObjectId][]*Client
// 	Register   chan *Client
// 	Unregister chan *Client
// }

// func NewHub() *Hub {
// 	tempMap := make(map[bson.ObjectId][]*Client)
// 	tempMap = getAllRoom(tempMap)
// 	return &Hub{
// 		Clients:    make(map[*Client]bool),
// 		Room:       tempMap,
// 		Register:   make(chan *Client),
// 		Unregister: make(chan *Client),
// 	}
// }

// // Run will execute Go Routines to check incoming Socket events
// func (hub *Hub) Run() {
// 	for {
// 		select {
// 		case client := <-hub.Register:
// 			HandleUserRegisterEvent(hub, client)

// 		case client := <-hub.Unregister:
// 			HandleUserDisconnectEvent(hub, client)

// 		}
// 	}
// }

// const (
// 	DBRoomName     = "Room"
// 	RoomCollection = "RoomData"
// )

// func getAllRoom(tempMap map[bson.ObjectId][]*Client) map[bson.ObjectId][]*Client {
// 	var rooms []model.Room
// 	var ConnectionDB, err = mgo.Dial(utills.MONGOENDPOINT)
// 	if err != nil {
// 		log.Println("error getAllRoom in NewHub()  ", err.Error())
// 	}
// 	err = ConnectionDB.DB(DBRoomName).C(RoomCollection).Find(nil).All(&rooms)

// 	for _, room := range rooms {
// 		for _, userID := range room.ListUser {
// 			tempMap[room.RoomID] = append(tempMap[room.RoomID], &Client{
// 				userID: userID,
// 			})
// 		}
// 	}
// 	return tempMap
// }
