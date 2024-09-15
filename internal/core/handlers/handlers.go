package handlers

import (
	"net/http"
	"time"

	"github.com/sounishnath003/jobprocessor/internal/core"
)

func HealthyHandler(w http.ResponseWriter, r *http.Request) {
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
	SendApiResponse(w, http.StatusOK, resp)
}

func HandleGetTasksList(w http.ResponseWriter, r *http.Request) {
	var (
		co = r.Context().Value("core").(*core.Core)
	)

	tasks := co.GetTasks()

	resp := NewApiResponse(http.StatusOK, nil, "get all jobs", tasks)
	SendApiResponse(w, http.StatusOK, resp)
}
