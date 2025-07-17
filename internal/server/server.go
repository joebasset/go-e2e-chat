package server

import (
	"github.com/joebasset/go-chat-e2e/internal/chat"
	"github.com/labstack/echo/v4"
)

func StartServer(hub *chat.Hub) {
	e := echo.New()
	createRoutes(e, hub)
	e.Logger.Fatal(e.Start(":8080"))
}
