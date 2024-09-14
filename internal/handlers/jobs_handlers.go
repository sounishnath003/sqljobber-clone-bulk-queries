package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sounishnath003/jobprocessor/internal/repositories"
)

func GetJobs(c echo.Context) error {
	jobs := repositories.GetJobs()

	resp := NewApiResponse(http.StatusOK, nil, "get all jobs", jobs)
	return c.JSON(http.StatusOK, resp)

}
