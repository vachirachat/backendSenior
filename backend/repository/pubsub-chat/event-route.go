package socket

import (
	"log"

	"github.com/globalsign/mgo/bson"
)

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

	case "disconnect room":
		log.Printf("Disconnect room Event triggered")
		selectedroomID := socketEventPayload.EventPayload.(map[string]interface{})["roomID"].(bson.ObjectId)
		BroadcastSocketEventToAllClientInRoom(client.hub, SocketEventStruct{
			EventName: socketEventPayload.EventName,
			EventPayload: RoomPayload{
				UserID: client.userID,
				RoomId: selectedroomID,
			},
		})

	case "join room":
		log.Printf("Connect room Event triggered")
		selectedroomID := socketEventPayload.EventPayload.(map[string]interface{})["roomID"].(bson.ObjectId)
		BroadcastSocketEventToAllClientInRoom(client.hub, SocketEventStruct{
			EventName: socketEventPayload.EventName,
			EventPayload: RoomPayload{
				UserID: client.userID,
				RoomId: selectedroomID,
			},
		})

	}
}
