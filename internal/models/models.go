package models

import "time"

type Job struct {
	ID     int    `json:"jobId"`
	Name   string `json:"name"`
	Ttl    int    `json:"ttl"`
	Status string `json:""`
}

const (
	StatusPending = "PENDING"
	StatusStarted = "STARTED"
	StatusFailure = "FAILURE"
	StatusSuccess = "SUCCESS"
	StatusRetry   = "RETRY"
)

type ResultBackend interface {
	NewResultSet(dbName, taskName string, ttl time.Duration) (ResultSet, error)
}

// ResultSet represents the set of results from an individual job that's executed.
type ResultSet interface {
}

// JobStatusResp - sends the REST response as a JOB status response
type JobStatusResp struct {
	JobID string `json:"job_id"`
	State string `json:"state"`
	Count int    `json:"count"`
	Error string `json:"error"`
}
