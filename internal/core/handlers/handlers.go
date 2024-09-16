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

// HandleGetTasksList returns the core.Core Tasks lists jobs supported
func HandleGetTasksList(w http.ResponseWriter, r *http.Request) {
	var (
		co = r.Context().Value("core").(*core.Core)
	)

	tasks := co.GetTasks()

	resp := NewApiResponse(http.StatusOK, nil, "lists of available tasks to submit as job", tasks)
	SendApiResponse(w, http.StatusOK, resp)
}

func HandleGetJobStatus(w http.ResponseWriter, r *http.Request) {
	var (
		co    = r.Context().Value("core").(*core.Core)
		jobID = r.PathValue("jobID")
	)

	out, err := co.GetJobStatus(jobID)
	if err != nil {
		SendApiResponse(w, http.StatusNotFound, NewApiResponse(http.StatusNotFound, err, "could not getjob status by jobID", nil))
		return
	}

	resp := NewApiResponse(http.StatusOK, nil, "jobstatus fetched successfully", out)
	SendApiResponse(w, http.StatusOK, resp)
}

func HandleGetPendingJobs(w http.ResponseWriter, r *http.Request) {
	var (
		co    = r.Context().Value("core").(*core.Core)
		queue = r.PathValue("queue")
	)

	out, err := co.GetPendingJobs(queue)
	if err != nil {
		SendApiResponse(w, http.StatusInternalServerError, NewApiResponse(http.StatusInternalServerError, err, "error fetching pending jobs", nil))
		return
	}

	resp := NewApiResponse(http.StatusOK, nil, "pending jobs information has been fetched", out)

	SendApiResponse(w, http.StatusOK, resp)
}
