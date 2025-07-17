package main

import (
	"github.com/joebasset/go-chat-e2e/internal/chat"
	"github.com/joebasset/go-chat-e2e/internal/server"
)

func main() {
	hub := chat.NewHub()
	go hub.Run()
	server.StartServer(hub)
}
