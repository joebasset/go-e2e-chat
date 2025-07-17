package chat

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// Hub maintains the set of active rooms and broadcasts messages to the
// rooms.
type Hub struct {
	// Registered rooms.
	rooms map[string]*Room
	// Register requests from the rooms.
	Register chan *Room

	// Unregister requests from rooms.
	unregister chan *Room
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for simplicity, but in production, restrict this.
		},
	}
)

func NewHub() *Hub {
	return &Hub{

		Register:   make(chan *Room),
		unregister: make(chan *Room),
		rooms:      make(map[string]*Room),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case room := <-h.Register:

			h.rooms[room.Id] = room
		case room := <-h.unregister:
			if _, ok := h.rooms[room.Id]; ok {
				delete(h.rooms, room.Id)
				close(room.send)
			}

		}
	}
}

func (h *Hub) WebSocketHandler(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	roomID := c.QueryParam("roomId")
	if roomID == "" {
		return c.String(http.StatusBadRequest, "Missing roomId")
	}

	room, ok := h.rooms[roomID]
	if !ok { // Room not found
		log.Printf("Room with ID '%s' not found.", roomID)
		// Optionally, send a WebSocket close message with a reason code
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Room not found"))
		return c.String(http.StatusNotFound, "Room not found") // Return HTTP status
	}

	go room.run()
	client := &Client{
		id:   uuid.NewString(),
		send: make(chan []byte, 256),
		conn: ws,
		room: room,
	}
	room.register <- client

	go client.readPump()
	go client.writePump()

	return nil
}
