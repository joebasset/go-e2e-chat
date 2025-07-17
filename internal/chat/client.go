package chat

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	id   string
	conn *websocket.Conn
	send chan []byte
	room *Room
}

func (c *Client) readPump() {
	// Ensure the client is unregistered from the room and connection is closed when this goroutine exits.
	defer func() {
		if c.room != nil { // Check if client is still associated with a room
			c.room.unregister <- c // Tell the room to unregister this client
		}
		c.conn.Close() // Close the WebSocket connection
	}()

	// Set read limits and pong handler for heartbeats
	c.conn.SetReadLimit(512)                                 // Max message size
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // Initial deadline
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // Reset deadline on pong
		return nil
	})

	for {

		// Read a message from the WebSocket.
		// messageType will be websocket.TextMessage or websocket.BinaryMessage
		// message is the actual byte content.
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("ReadMessage error for client %s: %v", c.id, err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Client %s WebSocket read error: %v", c.id, err)
			}
			break // Exit the read loop on error (connection closed)
		}

		// Process the incoming message.
		// For a chat app, you'll likely want to broadcast text messages.
		if messageType == websocket.TextMessage {
			// Prepend username or client ID before broadcasting
			fullMessage := []byte(c.id + ": " + string(message))
			fmt.Printf("MSG: %s", fullMessage)
			if c.room != nil {
				c.room.send <- fullMessage // Send the message to the room's broadcast channel
			} else {
				log.Printf("Client %s sent message but is not in a room: %s", c.id, string(message))
				// Optionally send an error back to client
				c.send <- []byte("Error: You are not in a room.")
			}
		}
		// You might handle other message types (e.g., binary for files) if needed.
	}
}

// writePump: Writes messages from the client's send channel to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second) // Send pings periodically (slightly less than pongWait)
	defer func() {
		ticker.Stop()
		c.conn.Close() // Close connection when this goroutine exits
	}()

	for {
		select {
		case message, ok := <-c.send:
			fmt.Printf("NEW MSG: %s", message)
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) // Set write deadline
			if !ok {
				// The client's send channel was closed by the room/hub.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return // Error getting writer, exit
			}
			w.Write(message)

			// Optionally, add more queued messages to the same WebSocket frame
			// (less common for simple chat, but can reduce overhead)
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'}) // Separator if multiple messages in one frame
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return // Error closing writer, exit
			}

		case <-ticker.C:
			// Send a ping message to keep the connection alive
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return // Ping failed, exit
			}
		}
	}
}
