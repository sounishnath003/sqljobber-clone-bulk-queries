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
