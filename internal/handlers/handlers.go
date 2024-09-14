package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func HealthyHandler(ctx echo.Context) error {
	generalInfo := map[string]string{
		"author":   "Sounish Nath",
		"Appname":  "jobprocessor",
		"Version":  "v0.1",
		"Hostname": "sounish-macbook-air-m1",
	}
	resp := NewApiResponse(http.StatusOK, nil, "API is running. Its healthly", generalInfo)
	return ctx.JSON(http.StatusOK, resp)
}
