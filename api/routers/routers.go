package routers

import (
	"lingolift/api/handler"

	"github.com/labstack/echo/v4"
)

func Load(e *echo.Echo) *echo.Echo {

	e.Static("/", "public")

	e.GET("/health", handler.Health)
	e.GET("/ws/assessment", handler.StreamAssessment)

	return e
}
