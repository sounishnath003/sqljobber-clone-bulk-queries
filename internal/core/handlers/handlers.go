package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/sounishnath003/jobprocessor/internal/core"
	"github.com/sounishnath003/jobprocessor/internal/models"
)

// reValidateName represents the character classes allowed in a job ID.
var reValidateName = regexp.MustCompile("(?i)^[a-z0-9-_:]+$")

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

// HandlePostJob submit a JOB for the mentioned {taskName} PathValue
func HandlePostJob(w http.ResponseWriter, r *http.Request) {
	var (
		co       = r.Context().Value("core").(*core.Core)
		taskName = r.PathValue("taskName")
	)

	if r.ContentLength == 0 {
		log.Println("request body is empty")
		SendApiResponse(w, http.StatusBadRequest, NewApiResponse(http.StatusBadRequest, errors.New("request body cannot be empty"), "request body is empty", nil))
		return
	}

	var (
		decoder = json.NewDecoder(r.Body)
		req     models.JobReq
	)

	if err := decoder.Decode(&req); err != nil {
		SendApiResponse(w, http.StatusBadRequest, NewApiResponse(http.StatusBadRequest, errors.New("error while parsing the request JSON"), "request JSON is not parsable as needed", nil))
		return
	}

	if !reValidateName.Match([]byte(req.JobID)) {
		SendApiResponse(w, http.StatusBadRequest, NewApiResponse(http.StatusBadRequest, errors.New("invalid characters in the `job_id`"), "invalid jobID", nil))
		return
	}

	// Create a job signature.
	out, err := co.NewJob(req, taskName)
	if err != nil {
		log.Println("could not create new job", "error", err, "task_name", taskName, "request", req)
		SendApiResponse(w, http.StatusInternalServerError, NewApiResponse(http.StatusInternalServerError, errors.New("could not create the job to process"), "job submission failed", nil))
		return
	}

	resp := NewApiResponse(http.StatusOK, nil, "job has been submitted successfully", out)
	SendApiResponse(w, http.StatusOK, resp)
}

// HandleGetJobStatus retrieves the status of a job by its ID.
func HandleGetJobStatus(w http.ResponseWriter, r *http.Request) {
	var (
		co    = r.Context().Value("core").(*core.Core) // Extracts the core instance from the request context.
		jobID = r.PathValue("jobID")                   // Extracts the job ID from the request path.
	)

	// Attempts to get the job status by the provided job ID.
	out, err := co.GetJobStatus(jobID)
	if err != nil {
		// If an error occurs, sends a response indicating the job status could not be fetched.
		SendApiResponse(w, http.StatusNotFound, NewApiResponse(http.StatusNotFound, err, "could not get job status by jobID", nil))
		return
	}

	// If the job status is successfully fetched, sends a response with the status.
	resp := NewApiResponse(http.StatusOK, nil, "job status fetched successfully", out)
	SendApiResponse(w, http.StatusOK, resp)
}

// HandleGetPendingJobs retrieves and sends the list of pending jobs for a given queue.
func HandleGetPendingJobs(w http.ResponseWriter, r *http.Request) {
	// Extracts the core instance from the request context.
	co := r.Context().Value("core").(*core.Core)
	// Extracts the queue name from the request path.
	queue := r.PathValue("queue")

	// Attempts to fetch the list of pending jobs for the specified queue.
	out, err := co.GetPendingJobs(queue)
	if err != nil {
		// If an error occurs, sends a response indicating the error fetching pending jobs.
		SendApiResponse(w, http.StatusInternalServerError, NewApiResponse(http.StatusInternalServerError, err, "error fetching pending jobs", nil))
		return
	}

	// Prepares a response with the fetched pending jobs information.
	resp := NewApiResponse(http.StatusOK, nil, "pending jobs information has been fetched", out)

	// Sends the response back to the client.
	SendApiResponse(w, http.StatusOK, resp)
}
