package socket

import (
	"backendSenior/model"
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// CreateNewSocketUser creates a new socket user
func CreateNewSocketUser(hub *model.Hub, connection *websocket.Conn, username string) {
	uniqueID := uuid.New()
	client := &model.Client{
		Hub:                 hub,
		WebSocketConnection: connection,
		Send:                make(chan model.SocketEventStruct),
		Username:            username,
		UserID:              uniqueID.String(),
	}

	go WritePump(client)
	go ReadPump(client)

	client.Hub.Register <- client
}

func ReadPump(c *model.Client) {
	var socketEventPayload model.SocketEventStruct

	defer unRegisterAndCloseConnection(c)

	setSocketPayloadReadConfig(c)

	for {
		_, payload, err := c.WebSocketConnection.ReadMessage()

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

func WritePump(c *model.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.WebSocketConnection.Close()
	}()
	for {
		select {
		case payload, ok := <-c.Send:
			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(payload)
			finalPayload := reqBodyBytes.Bytes()

			c.WebSocketConnection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.WebSocketConnection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.WebSocketConnection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(finalPayload)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				json.NewEncoder(reqBodyBytes).Encode(<-c.Send)
				w.Write(reqBodyBytes.Bytes())
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.WebSocketConnection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WebSocketConnection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func unRegisterAndCloseConnection(c *model.Client) {
	c.Hub.Unregister <- c
	c.WebSocketConnection.Close()
}

func setSocketPayloadReadConfig(c *model.Client) {
	c.WebSocketConnection.SetReadLimit(maxMessageSize)
	c.WebSocketConnection.SetReadDeadline(time.Now().Add(pongWait))
	c.WebSocketConnection.SetPongHandler(func(string) error { c.WebSocketConnection.SetReadDeadline(time.Now().Add(pongWait)); return nil })
}
