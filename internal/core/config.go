package core

import (
	"sync"
	"time"

	"github.com/sounishnath003/jobprocessor/internal/database"
)

// ConfigOpts construct to hold the necessarry ConfigOpts configuration throughout the systems
type ConfigOpts struct {
	PORT int
}

// Options construct to hold the necessarry option configuration throughout the systems
type Options struct {
	DefaultQueue            string
	DefaultGroupConcurrency int
	DefaultJobTTL           time.Duration
}

// Core construct to hold the necessarr configuration throughout the systems
type Core struct {
	Conf ConfigOpts
	Opts Options

	sourceDBs      database.Pool
	resultBackends ResultBackends
	mu             sync.RWMutex
}
