package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func HealthyHandler(ctx echo.Context) error {
	generalInfo := map[string]interface{}{
		"author":         "Sounish Nath",
		"appname":        "jobprocessor",
		"backendVersion": "v0.1",
		"apiVersioning": []string{
			"/api/v1",
			"/api/v2",
		},
		"releasedDate": time.Now().Format(time.DateOnly),
		"hostname":     "sounish-macbook-air-m1",
	}
	resp := NewApiResponse(http.StatusOK, nil, "API is running. Its healthly", generalInfo)
	return ctx.JSON(http.StatusOK, resp)
}
