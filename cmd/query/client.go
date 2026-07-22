package main

import (
	"time"

	"github.com/gofiber/contrib/v3/websocket"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// writePump(): This function waits for data in the client.send channel
// whenever the Hub sends a log to this channel,
// writePump retrieves it and sends it directly to the WebSocket for the browser to display
func (c *Client) writePump() {
	defer func() {
		_ = c.conn.Close()
	}()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// readPump(): This function is responsible for reading messages
// sent from the browser (maintaining a live connection/ping-pong)
// when the user closes the tab, it sends a signal to hub.unregister
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}
