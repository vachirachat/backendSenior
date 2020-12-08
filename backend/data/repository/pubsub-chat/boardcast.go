package repository

// import (
// 	"log"
// )

// // BroadcastSocketEventToAllClient will emit the socket events to all socket users
// func BroadcastSocketEventToAllClient(hub *Hub, payload SocketEventStruct) {
// 	log.Println("BroadcastSocketEventToAllClient hub.Clients", hub.Clients)
// 	for client := range hub.Clients {
// 		select {
// 		case client.send <- payload:
// 			log.Println("BroadcastSocketEventToAllClient payload", payload)
// 		default:
// 			close(client.send)
// 			delete(hub.Clients, client)
// 		}
// 	}
// }
