package upstream

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"proxySenior/domain/interface/repository"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pingPeriod = 20 * time.Second // NOTE: must be set according to server expectation
)

// UpstreamRepository is the client to upstream (controller)
// It manage websocket connection automatically
type UpstreamRepository struct {
	origin      string
	sendChannel chan []byte
	receivers   []chan []byte
	clientID    string
}

var _ repository.UpstreamMessageRepository = (*UpstreamRepository)(nil)

// NewUpStreamController create new upstream controller
func NewUpStreamController(origin string, clientID string) *UpstreamRepository {
	ctrl := &UpstreamRepository{
		origin:      origin,
		sendChannel: make(chan []byte, 10),
		receivers:   make([]chan []byte, 0),
		clientID:    clientID,
	}
	go ctrl.connect()
	return ctrl
}

func (upstream *UpstreamRepository) connect() {
	for {
		url := url.URL{
			Scheme:   "ws",
			Host:     upstream.origin,
			Path:     "/api/v1/chat/ws",
			RawQuery: "clientID=" + upstream.clientID,
		}

		c, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
		if err != nil {
			fmt.Println("Error connecting to upstream:", err)
			fmt.Println("Retrying in 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}
		fmt.Println("Connected to upstream")

		// is used to signal close of connection
		connCloseChan := make(chan struct{})

		go readPump(c, connCloseChan, upstream.receivers)
		go writePump(c, connCloseChan, upstream.sendChannel)

		<-connCloseChan

		fmt.Println("Reconnecting")
		c.Close()
	}
}

func readPump(conn *websocket.Conn, closeChan chan struct{}, receivers []chan []byte) {
	defer func() {
		conn.Close()
		close(closeChan)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1))
		for _, recv := range receivers {
			// TODO this might be bad ?
			go func() {
				recv <- message
			}()
		}
	}
}

func writePump(conn *websocket.Conn, cloesChannel chan struct{}, sendChannel chan []byte) {
	t := time.NewTicker(pingPeriod)
	defer func() {
		conn.Close()
		t.Stop()
	}()

	for {
		select {
		case _, ok := <-cloesChannel:
			if !ok {
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				// The hub closed the channel.
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
		case message := <-sendChannel:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			conn.WriteMessage(websocket.TextMessage, message)

			// Add queued chat messages to the current websocket message.
			n := len(sendChannel)
			for i := 0; i < n; i++ {
				conn.WriteMessage(websocket.TextMessage, <-sendChannel)
			}

		case <-t.C: // ping
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

//SendMessage sends message to upstream channel
func (upstream *UpstreamRepository) SendMessage(message []byte) error {
	upstream.sendChannel <- message
	return nil
}

// RegisterHandler add channel to receive message it will send message to that channel
func (upstream *UpstreamRepository) RegisterHandler(channel chan []byte) error {
	for _, r := range upstream.receivers {
		if r == channel {
			return errors.New("Channel Already Exists")
		}
	}
	upstream.receivers = append(upstream.receivers, channel)
	return nil
}

// UnRegisterHandler unregister channel for receiving messageBinaryMessage
func (upstream *UpstreamRepository) UnRegisterHandler(channel chan []byte) error {
	n := len(upstream.receivers)
	for i := 0; i < n; i++ {
		if channel == upstream.receivers[i] {
			upstream.receivers[i], upstream.receivers[n-1] = upstream.receivers[n-1], upstream.receivers[i]
			upstream.receivers = upstream.receivers[:n-1]
			return nil
		}
	}
	return errors.New("No Channel Removed")
}
