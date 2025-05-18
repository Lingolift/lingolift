package handler

import "github.com/labstack/echo/v4"

// Health
func Health(c echo.Context) error {
	return c.JSON(200, map[string]bool{
		"health": true,
	})
}
