package core

import (
	"sync"
	"time"

	"github.com/kalbhor/tasqueue/v2"
	"github.com/sounishnath003/jobprocessor/internal/database"
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

	q *tasqueue.Server
}
