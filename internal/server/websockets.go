package server

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/joebasset/go-chat-e2e/internal/chat"
	"github.com/labstack/echo/v4"

	"golang.org/x/net/websocket"
)

var rooms = make(map[string][]*chat.Client)

func addClientToRoom(roomID string, ws *websocket.Conn) {
	newClient := chat.Client{
		Id:   uuid.NewString(),
		Conn: ws,
	}
	rooms[roomID] = append(rooms[roomID], &newClient)
}

func removeClientFromRoom(roomID string, ws *websocket.Conn) {
	clients := rooms[roomID]
	for i, c := range clients {
		if c.Conn == ws {
			// Remove client from slice
			fmt.Printf("Removing Client: %s", c.Id)
			rooms[roomID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	// Optional: delete room if empty
	if len(rooms[roomID]) == 0 {
		delete(rooms, roomID)
	}
}
func broadcastToRoom(roomID string, ws *websocket.Conn, message string) {
	for _, client := range rooms[roomID] {

		if client.Conn == ws {
			// skip the sender
			continue
		}
		err := websocket.Message.Send(client.Conn, message+client.Id)
		if err != nil {
			// handle error, maybe remove client
			client.Conn.Close()
			removeClientFromRoom(roomID, client.Conn)
		}
	}
}

func CreateConnection(c echo.Context) error {
	roomID := c.QueryParam("roomId")
	if roomID == "" {
		return c.String(http.StatusBadRequest, "Missing roomId")
	}

	wsHandler := func(ws *websocket.Conn) {
		defer ws.Close()

		// Add client to room
		if rooms[roomID] == nil {
			rooms[roomID] = []*chat.Client{}
		}
		addClientToRoom(roomID, ws)

		for {
			var msg string
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				break
			}
			broadcastToRoom(roomID, ws, msg)
		}

		removeClientFromRoom(roomID, ws)
	}

	websocket.Handler(wsHandler).ServeHTTP(c.Response(), c.Request())
	return nil
}
