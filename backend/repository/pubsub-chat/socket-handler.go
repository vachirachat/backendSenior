package socket

import (
	"backendSenior/repository"
	"log"
	"time"

	"github.com/globalsign/mgo/bson"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
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
func HandleRoomRegisterEvent(hub *Hub, client *Client) {
	_, ok := hub.Clients[client]
	if ok {
		// Query UserRoom Data from user Database ->

		roomlist, err := repository.GetRoomWithUserID(client.userID)
		if err != nil {
			log.Println(err)
		}

		for _, room := range roomlist {
			hub.Room[room] = append(hub.Room[room], client)
		}

		handleSocketPayloadEvents(client, SocketEventStruct{
			EventName:    "connect room",
			EventPayload: client.userID,
		})
	}
}

// HandleUserDisconnectEvent will handle the Disconnect. event for socket users
func HandleRoomDisconnectEvent(hub *Hub, client *Client) {
	// _, ok := hub.Clients[client]
	// if ok {
	// 	delete(hub.Clients, client)
	// 	close(client.send)

	// 	handleSocketPayloadEvents(client, SocketEventStruct{
	// 		EventName:    "disconnect room",
	// 		EventPayload: client.userID,
	// 	})
	// }
}

// BroadcastSocketEventToAllClient will emit the socket events to all socket users
func BroadcastSocketEventToAllClient(hub *Hub, payload SocketEventStruct) {
	log.Println("BroadcastSocketEventToAllClient hub.Clients", hub.Clients)
	for client := range hub.Clients {
		select {
		case client.send <- payload:
			log.Println("BroadcastSocketEventToAllClient payload", payload)
		default:
			close(client.send)
			delete(hub.Clients, client)
		}
	}
}

func handleSocketPayloadEvents(client *Client, socketEventPayload SocketEventStruct) {
	var socketEventResponse SocketEventStruct
	switch socketEventPayload.EventName {
	case "join":
		log.Printf("Join Event triggered")
		BroadcastSocketEventToAllClient(client.hub, SocketEventStruct{
			EventName: socketEventPayload.EventName,
			EventPayload: JoinDisconnectPayload{
				UserID: client.userID,
				Users:  getAllConnectedUsers(client.hub),
			},
		})

	case "disconnect":
		log.Printf("Disconnect Event triggered")
		BroadcastSocketEventToAllClient(client.hub, SocketEventStruct{
			EventName: socketEventPayload.EventName,
			EventPayload: JoinDisconnectPayload{
				UserID: client.userID,
				Users:  getAllConnectedUsers(client.hub),
			},
		})

	case "message":
		log.Printf("Message Clients Event triggered")

		/*
			 JSON

			 {
				 username
				 message
				 userID
				 roomID
			 }
		*/

		selectedUserID := socketEventPayload.EventPayload.(map[string]interface{})["userID"].(bson.ObjectId)
		selectedroomID := socketEventPayload.EventPayload.(map[string]interface{})["roomID"].(bson.ObjectId)

		socketEventResponse.EventName = "message response"
		socketEventResponse.EventPayload = map[string]interface{}{
			"username": getUsernameByUserID(client.hub, selectedUserID),
			"message":  socketEventPayload.EventPayload.(map[string]interface{})["message"],
			"userID":   selectedUserID,
		}
		EmitToSpecificClient(client.hub, socketEventResponse, selectedUserID, selectedroomID)
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
