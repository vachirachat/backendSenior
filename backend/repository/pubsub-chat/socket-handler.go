package socket

import (
	"backendSenior/model"
	"log"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// HandleUserRegisterEvent will handle the Join event for New socket users
func HandleUserRegisterEvent(hub *model.Hub, client *model.Client) {
	hub.Clients[client] = true
	handleSocketPayloadEvents(client, model.SocketEventStruct{
		EventName:    "join",
		EventPayload: client.UserID,
	})
}

// HandleUserDisconnectEvent will handle the Disconnect. event for socket users
func HandleUserDisconnectEvent(hub *model.Hub, client *model.Client) {
	_, ok := hub.Clients[client]
	if ok {
		delete(hub.Clients, client)
		close(client.Send)

		handleSocketPayloadEvents(client, model.SocketEventStruct{
			EventName:    "disconnect",
			EventPayload: client.UserID,
		})
	}
}

// EmitToSpecificClient will emit the socket event to specific socket user
func EmitToSpecificClient(hub *model.Hub, payload model.SocketEventStruct, userID string) {
	for client := range hub.Clients {
		if client.UserID == userID {
			select {
			case client.Send <- payload:
			default:
				close(client.Send)
				delete(hub.Clients, client)
			}
		}
	}
}

// BroadcastSocketEventToAllClient will emit the socket events to all socket users
func BroadcastSocketEventToAllClient(hub *model.Hub, payload model.SocketEventStruct) {
	log.Println(hub.Clients)
	for client := range hub.Clients {
		select {
		case client.Send <- payload:
			log.Println(payload)
		default:
			close(client.Send)
			delete(hub.Clients, client)
		}
	}
}

func handleSocketPayloadEvents(client *model.Client, socketEventPayload model.SocketEventStruct) {
	var socketEventResponse model.SocketEventStruct
	switch socketEventPayload.EventName {
	case "join":
		log.Printf("Join Event triggered")
		BroadcastSocketEventToAllClient(client.Hub, model.SocketEventStruct{
			EventName: socketEventPayload.EventName,
			EventPayload: model.JoinDisconnectPayload{
				UserID: client.UserID,
				Users:  getAllConnectedUsers(client.Hub),
			},
		})

	case "disconnect":
		log.Printf("Disconnect Event triggered")
		BroadcastSocketEventToAllClient(client.Hub, model.SocketEventStruct{
			EventName: socketEventPayload.EventName,
			EventPayload: model.JoinDisconnectPayload{
				UserID: client.UserID,
				Users:  getAllConnectedUsers(client.Hub),
			},
		})

	case "message":
		log.Printf("Message Event triggered")
		selectedUserID := socketEventPayload.EventPayload.(map[string]interface{})["userID"].(string)
		socketEventResponse.EventName = "message response"
		socketEventResponse.EventPayload = map[string]interface{}{
			"username": getUsernameByUserID(client.Hub, selectedUserID),
			"message":  socketEventPayload.EventPayload.(map[string]interface{})["message"],
			"userID":   selectedUserID,
		}
		EmitToSpecificClient(client.Hub, socketEventResponse, selectedUserID)
	}
}

func getUsernameByUserID(hub *model.Hub, userID string) string {
	var username string
	for client := range hub.Clients {
		if client.UserID == userID {
			username = client.Username
		}
	}
	return username
}

func getAllConnectedUsers(hub *model.Hub) []model.UserStruct {
	var users []model.UserStruct
	for singleClient := range hub.Clients {
		users = append(users, model.UserStruct{
			Username: singleClient.Username,
			UserID:   singleClient.UserID,
		})
	}
	return users
}
