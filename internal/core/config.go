package core

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"math/rand"

	"github.com/kalbhor/tasqueue/v2"
	"github.com/sounishnath003/jobprocessor/internal/database"
	"github.com/sounishnath003/jobprocessor/internal/models"
)

type RedisBroker struct {
	Addrs        []string
	Password     string
	DB           int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type RedisResult struct {
	RedisBroker
	Expiry     time.Duration
	MetaExpiry time.Duration
}

// ConfigOpts construct to hold the necessarry ConfigOpts configuration throughout the systems
type ConfigOpts struct {
	PORT        int
	BrokerQueue RedisBroker
	ResultQueue RedisResult
}

// Options construct to hold the necessarry option configuration throughout the systems
type Options struct {
	DefaultQueue            string
	DefaultGroupConcurrency int
	DefaultJobTTL           time.Duration

	// DSNs for connecting to the broker backend and the broker	state backend.
	Broker  tasqueue.Broker
	Results tasqueue.Results
}

// Core construct to hold the necessarr configuration throughout the systems
type Core struct {
	Conf ConfigOpts
	Opts Options

	sourceDBs      database.Pool
	resultBackends ResultBackends
	mu             sync.RWMutex

	q      *tasqueue.Server
	tasks  Tasks
	jobCtx map[string]context.CancelFunc
	lo     *slog.Logger
}

type Task struct {
	Name           string         `json:"name"`
	Queue          string         `json:"queue"`
	Conc           int            `json:"concurrency"`
	Stmt           *sql.Stmt      `json:"-"`
	Raw            string         `json:"raw,omitempty"`
	DBs            database.Pool  `json:"-"`
	ResultBackends ResultBackends `json:"-"`
}

// Tasks represents a map of prepared SQL statements.
type Tasks map[string]Task

// GetRandom returns a random results backend from the map.
func (r ResultBackends) GetRandom() (string, models.ResultBackend) {
	stop := 0
	if len(r) > 1 {
		stop = rand.Intn(len(r))
	}

	i := 0
	for name, v := range r {
		if i == stop {
			return name, v
		}

		i++
	}

	// This'll never happen.
	return "", nil
}
