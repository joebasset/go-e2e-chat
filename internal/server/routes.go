package server

import (
	"net/http"

	"github.com/joebasset/go-chat-e2e/internal/chat"
	"github.com/labstack/echo/v4"
)

func createNewRoom(c echo.Context, hub *chat.Hub) error {

	newRoom := chat.NewRoom()
	hub.Register <- newRoom

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": true,
		"id":     newRoom.Id,
	})

}
func createRoutes(e *echo.Echo, hub *chat.Hub) {
	e.GET("/", func(c echo.Context) error {
		return c.File("../../web/index.html")
	})
	e.GET("/ws", hub.WebSocketHandler)

	room := e.Group("/rooms")

	room.POST("/create", func(c echo.Context) error {
		createNewRoom(c, hub)
		return nil
	})

}
