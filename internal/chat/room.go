package chat

import (
	"fmt"

	"github.com/google/uuid"
)

type Room struct {
	Id         string
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	send       chan []byte
}

func NewRoom() *Room {
	return &Room{
		Id:         uuid.NewString(),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		send:       make(chan []byte),
	}
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.register:

			r.clients[client.id] = client

		case client := <-r.unregister:
			{
				fmt.Printf("DELETING CLIENT: %s", client.id)
				delete(r.clients, client.id)
				close(client.send)
			}

		case message := <-r.send:
			{
				for _, client := range r.clients {
					select {
					case client.send <- message:
					default:

						close(client.send)
						delete(r.clients, client.id)
					}
				}
			}

		}
	}
}
