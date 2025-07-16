package server

import (
	"github.com/labstack/echo/v4"
)

func Server() {
	e := echo.New()
	CreateRoutes(e)
	e.Logger.Fatal(e.Start(":8080"))
}
