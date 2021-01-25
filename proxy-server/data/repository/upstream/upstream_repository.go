package upstream

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
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
	origin           string
	sendChannel      chan []byte
	receivers        []chan []byte
	onConnectRecv    []chan struct{}
	onDisconnectRecv []chan struct{}
	stopChan         chan struct{}
	clientID         string
	clientSecret     string
}

var _ repository.UpstreamMessageRepository = (*UpstreamRepository)(nil)

// NewUpStreamController create new upstream controller
func NewUpStreamController(origin string, clientID string, clientSecret string) *UpstreamRepository {
	ctrl := &UpstreamRepository{
		origin:           origin,
		sendChannel:      make(chan []byte, 10),
		receivers:        make([]chan []byte, 0),
		onConnectRecv:    make([]chan struct{}, 0),
		onDisconnectRecv: make([]chan struct{}, 0),
		stopChan:         make(chan struct{}),
		clientID:         clientID,
		clientSecret:     clientSecret,
	}
	go ctrl.connect()
	return ctrl
}

// Stop disconnect and stop controller
func (upstream *UpstreamRepository) Stop() {
	close(upstream.stopChan)
}

func (upstream *UpstreamRepository) connect() {
loop:
	for {
		url := url.URL{
			Scheme: "ws",
			Host:   upstream.origin,
			Path:   "/api/v1/chat/ws",
		}

		authHeader := base64.StdEncoding.EncodeToString([]byte(upstream.clientID + ":" + upstream.clientSecret))

		var h = http.Header{}
		h.Add("Authorization", "Basic "+authHeader)

		c, _, err := websocket.DefaultDialer.Dial(url.String(), h)
		if err != nil {
			fmt.Println("Error connecting to upstream:", err)
			fmt.Println("Retrying in 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}
		fmt.Println("Connected to upstream")

		for _, c := range upstream.onConnectRecv {
			select {
			case c <- struct{}{}:
			default:
			}
		}

		// is used to signal close of connection
		connCloseChan := make(chan struct{})

		go readPump(c, connCloseChan, upstream.receivers)
		go writePump(c, connCloseChan, upstream.sendChannel)

		select {
		case <-connCloseChan:
			fmt.Print("conncetion closed")
		case <-upstream.stopChan:
			break loop
		}
		fmt.Println("Reconnecting")
		c.Close()
	}
	for _, c := range upstream.onDisconnectRecv {
		select {
		case c <- struct{}{}:
		default:
		}
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
			select {
			case recv <- message:
			default:
			}
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

// OnConnect register channel to be notified when upstream is connected
func (upstream *UpstreamRepository) OnConnect(channel chan struct{}) {
	upstream.onConnectRecv = append(upstream.onConnectRecv, channel)
}

// OffConnect unregister channel from being notified when upstream is connected
func (upstream *UpstreamRepository) OffConnect(channel chan struct{}) {
	idx := 0
	for i := 0; i < len(upstream.onConnectRecv); i++ {
		if upstream.onConnectRecv[i] != channel {
			upstream.onConnectRecv[idx] = upstream.onConnectRecv[i]
			idx++
		}
	}
	upstream.onConnectRecv = upstream.onConnectRecv[:idx]
}

// OnDisconnect register channel to be notified when upstream is disconnected
func (upstream *UpstreamRepository) OnDisconnect(channel chan struct{}) {
	upstream.onDisconnectRecv = append(upstream.onDisconnectRecv, channel)
}

// OffDisconnect unregister channel from being notified when upstream is disconnected
func (upstream *UpstreamRepository) OffDisconnect(channel chan struct{}) {
	idx := 0
	for i := 0; i < len(upstream.onDisconnectRecv); i++ {
		if upstream.onDisconnectRecv[i] != channel {
			upstream.onDisconnectRecv[idx] = upstream.onDisconnectRecv[i]
			idx++
		}
	}
	upstream.onDisconnectRecv = upstream.onDisconnectRecv[:idx]
}
