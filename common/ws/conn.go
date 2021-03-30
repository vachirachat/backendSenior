package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/reactivex/rxgo/v2"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

// writeCmd is write-like method as defined by gorilla
// it support close or send data
type writeCmd struct {
	close bool
	data  []byte
	resp  chan bool
}

type Connection struct {
	conn         *websocket.Conn
	closed       bool
	isStarted    bool
	writeChannel chan writeCmd
	readChannel  chan rxgo.Item // convenient for rxgo
	//
	obs rxgo.Observable
}

func FromConnection(conn *websocket.Conn) *Connection {
	readChan := make(chan rxgo.Item, 100)
	c := &Connection{
		conn:         conn,
		closed:       false,
		isStarted:    false,
		writeChannel: make(chan writeCmd, 100),
		readChannel:  readChan,
		obs:          rxgo.FromEventSource(readChan),
	}
	return c
}

func (c *Connection) StartLoop() {
	if !c.isStarted {
		c.isStarted = true
		go c.readLoop()
		go c.writeLoop()
	} else {
		panic("not allowed to start loop more than one time")
	}
}

// readLoop read message and pipe to channel
func (c *Connection) readLoop() {
	defer func() {
		c.closed = true
		close(c.readChannel) // imply that observable is closed too
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil || c.closed {
			break
		}
		fmt.Printf("received message %s\n", data)
		c.readChannel <- rxgo.Of(data)
	}
}

func (c *Connection) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.closed = true

		n := len(c.writeChannel)
		for i := 0; i < n; i++ {
			cmd := <-c.writeChannel
			if cmd.close { // close always success
				cmd.resp <- true
			} else { // send failed coz channel is closed
				cmd.resp <- false
			}

		}
	}()

	for {
		select {
		case cmd := <-c.writeChannel: // we never close the channel
			if cmd.close { // manually closed, send close message, and bye
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				cmd.resp <- true
				break
			} else if c.closed { // force closed, just stop
				break
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, cmd.data); err != nil {
				cmd.resp <- false // error
				break
			}
			cmd.resp <- true
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				break
			}
		}

	}

}

// TODO: ensure no send to closed channel
func (c *Connection) Send(message []byte) error {
	respChan := make(chan bool, 1)
	c.writeChannel <- writeCmd{data: message, resp: respChan}
	if ok := <-respChan; ok {
		return nil
	}
	return errors.New("send error, connection closed")
}

func (c *Connection) SendJSON(data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	return c.Send(b)
}

// Close close the connection
// closing already closed is no-op
func (c *Connection) Close() {
	c.writeChannel <- writeCmd{close: true}
}

// Observable return observable
// Please do not .Observe() it
func (c *Connection) Observable() rxgo.Observable {
	return c.obs
}
