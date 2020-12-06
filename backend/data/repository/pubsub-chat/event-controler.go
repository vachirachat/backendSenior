package repository

import (
	"log"
	"time"

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

	case "join room":
		log.Printf("join room Event triggered")
		selectedroomID := socketEventPayload.EventPayload.(map[string]interface{})["roomID"].(bson.ObjectId)
		selecteduserID := socketEventPayload.EventPayload.(map[string]interface{})["userID"].(bson.ObjectId)

		socketEventResponse.EventName = "join room response"
		socketEventResponse.EventPayload = map[string]interface{}{
			"roomID":    selectedroomID,
			"userID":    selecteduserID,
			"timestamp": socketEventPayload.EventPayload.(map[string]interface{})["timestamp"].(time.Time),
		}
		HandleJoinRoomEvent(client.hub, socketEventResponse, selectedroomID, client)

	case "disconnect room":
		log.Printf("leave room Event triggered")
		selectedroomID := socketEventPayload.EventPayload.(map[string]interface{})["roomID"].(bson.ObjectId)
		selecteduserID := socketEventPayload.EventPayload.(map[string]interface{})["userID"].(bson.ObjectId)

		socketEventResponse.EventName = "join room response"
		socketEventResponse.EventPayload = map[string]interface{}{
			"roomID":    selectedroomID,
			"userID":    selecteduserID,
			"timestamp": socketEventPayload.EventPayload.(map[string]interface{})["timestamp"].(time.Time),
		}
		HandleJoinRoomEvent(client.hub, socketEventResponse, selectedroomID, client)

	case "message private":
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

		selectedroomID := socketEventPayload.EventPayload.(map[string]interface{})["roomID"].(bson.ObjectId)
		socketEventResponse.EventName = "message private response"
		socketEventResponse.EventPayload = map[string]interface{}{
			"message":   socketEventPayload.EventPayload.(map[string]interface{})["message"],
			"roomID":    selectedroomID,
			"userID":    socketEventPayload.EventPayload.(map[string]interface{})["userID"],
			"timestamp": socketEventPayload.EventPayload.(map[string]interface{})["timestamp"].(time.Time),
		}
		EmitToMessage(client.hub, socketEventResponse, selectedroomID, "PRIVATE")

	case "message group":
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
		msgPayload := socketEventPayload.EventPayload.(messagePayload)
		selectedroomID := msgPayload.RoomId

		socketEventResponse.EventName = "message group response"
		socketEventResponse.EventPayload = map[string]interface{}{
			"message":   msgPayload.Message,
			"userID":    msgPayload.UserId,
			"roomID":    msgPayload.RoomId,
			"timestamp": msgPayload.Timestamp,
		}
		EmitToMessage(client.hub, socketEventResponse, selectedroomID, "GROUP")

	}
}
