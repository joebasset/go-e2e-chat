package chat

type Room struct {
	Id         string
	Clients    []Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			r.Clients[client.Id] = client

		case client := <-r.Unregister:
			delete(r.Clients, client.Id)
			close(client.Send)

		case message := <-r.Broadcast:
			for _, client := range r.Clients {
				select {
				case client.Send <- message:
				default:
					// if client buffer is full, disconnect it
					close(client.Send)
					delete(r.Clients, client.Id)
				}
			}
		}
	}
}
