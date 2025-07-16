package chat

import "golang.org/x/net/websocket"

type Client struct {
	Id   string
	Conn *websocket.Conn
	Send chan []byte
	Room *Room
}
