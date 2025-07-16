package server

import (
	"github.com/joebasset/go-chat-e2e/internal/chat"
	"github.com/labstack/echo/v4"
)

func CreateRoutes(e *echo.Echo) {
	e.GET("/", func(c echo.Context) error {
		return c.File("../../web/index.html")
	})
	e.GET("/ws", CreateConnection)

	room := e.Group("/rooms")

	room.POST("/create", chat.CreateRoomHandler)

}
