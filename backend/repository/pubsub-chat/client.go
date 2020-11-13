package socket

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/websocket"
)

func unRegisterAndCloseConnection(c *Client) {
	c.hub.Unregister <- c
	c.webSocketConnection.Close()
}

func setSocketPayloadReadConfig(c *Client) {
	c.webSocketConnection.SetReadLimit(maxMessageSize)
	c.webSocketConnection.SetReadDeadline(time.Now().Add(pongWait))
	c.webSocketConnection.SetPongHandler(func(string) error { c.webSocketConnection.SetReadDeadline(time.Now().Add(pongWait)); return nil })
}

// CreateNewSocketUser creates a new socket user
func CreateNewSocketUser(hub *Hub, connection *websocket.Conn, userID bson.ObjectId, username string, room []bson.ObjectId) {
	//uniqueID := uuid.New()
	client := &Client{
		hub:                 hub,
		webSocketConnection: connection,
		send:                make(chan SocketEventStruct),
		username:            username,
		userID:              userID,
		Room:                room,
	}

	go client.WritePump()
	go client.ReadPump()

	client.hub.Register <- client

}

func (c *Client) ReadPump() {
	var socketEventPayload SocketEventStruct

	defer unRegisterAndCloseConnection(c)

	setSocketPayloadReadConfig(c)

	for {
		_, payload, err := c.webSocketConnection.ReadMessage()

		decoder := json.NewDecoder(bytes.NewReader(payload))
		decoderErr := decoder.Decode(&socketEventPayload)

		if decoderErr != nil {
			log.Printf("error: %v", decoderErr)
			break
		}

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error ===: %v", err)
			}
			break
		}

		handleSocketPayloadEvents(c, socketEventPayload)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.webSocketConnection.Close()
	}()
	for {
		select {
		case payload, ok := <-c.send:
			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(payload)
			finalPayload := reqBodyBytes.Bytes()

			c.webSocketConnection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.webSocketConnection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.webSocketConnection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(finalPayload)

			n := len(c.send)
			for i := 0; i < n; i++ {
				json.NewEncoder(reqBodyBytes).Encode(<-c.send)
				w.Write(reqBodyBytes.Bytes())
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.webSocketConnection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.webSocketConnection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
