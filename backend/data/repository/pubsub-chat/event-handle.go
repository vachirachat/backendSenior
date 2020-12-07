package repository

import (
	repository "backendSenior/data/repository/mongo_repository"

	"backendSenior/domain/model"
	"backendSenior/utills"
	"log"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// HandleUserRegisterEvent will handle the Join event for New socket users
func HandleUserRegisterEvent(hub *Hub, client *Client) {
	hub.Clients[client] = true

	//load user to hub
	for _, room := range client.Room {
		for i, cli := range hub.Room[room] {
			if client.userID == cli.userID {
				hub.Room[room][i] = client
			}
		}
	}

	log.Println("HandleUserRegisterEvent", client.username)
	handleSocketPayloadEvents(client, SocketEventStruct{
		EventName:    "join",
		EventPayload: client.userID,
	})
}

// HandleUserDisconnectEvent will handle the Disconnect. event for socket users
func HandleUserDisconnectEvent(hub *Hub, client *Client) {
	_, ok := hub.Clients[client]
	if ok {
		delete(hub.Clients, client)
		close(client.send)

		handleSocketPayloadEvents(client, SocketEventStruct{
			EventName:    "disconnect",
			EventPayload: client.userID,
		})
	}
}

// HandleJoinRoomEvent will handle the Join event for users
func HandleJoinRoomEvent(hub *Hub, payload SocketEventStruct, roomID bson.ObjectId, client *Client) {
	_, ok1 := hub.Clients[client]
	_, ok2 := hub.Room[roomID]
	if ok1 && ok2 {
		hub.Room[roomID] = append(hub.Room[roomID], client)
		//mgo add user from room
		//TODO

		select {
		case client.send <- payload:
			err := repository.AddMemberToRoom(roomID, client.userID)
			if err != nil {
				log.Println("error HandleJoinRoomEvent to DB", err.Error())
				return
			}
		default:
			close(client.send)
			delete(hub.Clients, client)

		}
	} else {
		log.Println("room not exist")
	}

}

// HandleLeaveRoomEvent will handle the leave event for users
func HandleLeaveRoomEvent(hub *Hub, payload SocketEventStruct, roomID bson.ObjectId, client *Client) {
	_, ok1 := hub.Clients[client]
	_, ok2 := hub.Room[roomID]
	if ok1 && ok2 {
		for i, clnt := range hub.Room[roomID] {
			if clnt == client {
				hub.Room[roomID] = append(hub.Room[roomID][:i], hub.Room[roomID][i+1:]...)
				//mgo delete user from room
				//TODO

				select {
				case client.send <- payload:
					err := repository.DeleteMemberToRoom(roomID, client.userID)
					if err != nil {
						log.Println("error HandleLeaveRoomEvent to DB", err.Error())
						return
					}
				default:
					close(client.send)
					delete(hub.Clients, client)
				}

			} else {
				log.Println("user dose not exist in room early")
			}
		}
	} else {
		log.Println("room not exist")
	}

}

// EmitToSpecificClient will emit the socket event to specific socket user
func EmitToMessage(hub *Hub, payload SocketEventStruct, room bson.ObjectId, FlagRoomTYPE string) {
	_, ok := hub.Room[room]
	if ok {
		for _, client := range hub.Room[room] {
			// Now send only online user
			if hub.Clients[client] {
				select {
				case client.send <- payload:
				default:
					close(client.send)
					delete(hub.Clients, client)

				}
			}
			//write DB message room
			sender := payload.EventPayload.(map[string]interface{})["userID"].(bson.ObjectId)
			username := getUsernameByUserID(hub, sender)
			addUserMessage(sender, room, payload, FlagRoomTYPE, username)
		}

	} else {
		log.Fatalln("No Room " + room + "in hub")
	}

}

func getUsernameByUserID(hub *Hub, userID bson.ObjectId) string {
	var username string
	for client := range hub.Clients {
		if client.userID == userID {
			username = client.username
		}
	}
	return username
}

func getAllConnectedUsers(hub *Hub) []UserStruct {
	var users []UserStruct
	for singleClient := range hub.Clients {
		users = append(users, UserStruct{
			Username: singleClient.username,
			UserID:   singleClient.userID,
		})
	}
	return users
}

// Temp data
const (
	DBMessage         = "Message"
	collectionMessage = "MessageData"
)

// DB get USer-room place
func addUserMessage(userId bson.ObjectId, roomId bson.ObjectId, payload SocketEventStruct, Type string, Name string) {
	var ConnectionDB, err = mgo.Dial(utills.MONGOENDPOINT)
	var message model.Message
	// insert new message
	message.RoomID = roomId
	message.UserID = userId
	message.Name = Name
	message.TimeStamp = payload.EventPayload.(map[string]interface{})["timestamp"].(time.Time)
	message.Data = payload.EventPayload.(map[string]interface{})["message"].(string)
	message.Type = Type

	if err != nil {
		log.Println("error addUserMessage", err.Error())
		return
	}
	err = ConnectionDB.DB(DBMessage).C(collectionMessage).Insert(message)
	if err != nil {
		log.Println("error addUserMessage mongo cant ADD", err.Error())
		return
	}
}
