package socket

import (
	"log"

	"github.com/globalsign/mgo/bson"
)

// HandleUserRegisterEvent will handle the Join event for New socket users
func HandleUserRegisterEvent(hub *Hub, client *Client) {
	hub.Clients[client] = true
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

// HandleUserRegisterEvent will handle the Join event for New socket users
func HandleInitConnectRegisterEvent(hub *Hub, client *Client) {
	_, ok := hub.Clients[client]
	if ok {

	}
}

// EmitToSpecificClient will emit the socket event to specific socket user
func EmitToSpecificClient(hub *Hub, payload SocketEventStruct, userID bson.ObjectId, room bson.ObjectId) {
	if room.String() == "" {
		for client := range hub.Clients {
			if client.userID == userID {
				select {
				case client.send <- payload:
				default:
					close(client.send)
					delete(hub.Clients, client)
				}
			}
		}
	} else {
		for _, client := range hub.Room[room] {
			if hub.Clients[client] {
				select {
				case client.send <- payload:
				default:
					close(client.send)
					delete(hub.Clients, client)

				}
			}

		}
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
