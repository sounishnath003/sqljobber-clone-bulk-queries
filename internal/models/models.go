package models

import (
	"database/sql"
	"time"
)

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
	RegisterColTypes([]string, []*sql.ColumnType) error
	IsColTypesRegistered() bool
	WriteCols([]string) error
	WriteRow([]interface{}) error
	Flush() error
	Close() error
}

// JobReq represents a job request.
type JobReq struct {
	TaskName string   `json:"task"`
	JobID    string   `json:"job_id"`
	Queue    string   `json:"queue"`
	ETA      string   `json:"eta"`
	Retries  int      `json:"retries"`
	TTL      int      `json:"ttl"`
	Args     []string `json:"args"`
	DB       string   `json:"db"`

	ttlDuration time.Duration
}

// JobResp is the response sent to a job request.
type JobResp struct {
	JobID    string     `json:"job_id"`
	TaskName string     `json:"task"`
	Queue    string     `json:"queue"`
	ETA      *time.Time `json:"eta"`
	Retries  int        `json:"retries"`
}

// JobStatusResp - sends the REST response as a JOB status response
type JobStatusResp struct {
	JobID string `json:"job_id"`
	State string `json:"state"`
	Count int    `json:"count"`
	Error string `json:"error"`
}
