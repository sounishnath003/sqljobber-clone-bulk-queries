package core

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kalbhor/tasqueue/v2"
	bredis "github.com/kalbhor/tasqueue/v2/brokers/redis"
	rredis "github.com/kalbhor/tasqueue/v2/results/redis"

	"github.com/knadh/goyesql/v2"

	"github.com/sounishnath003/jobprocessor/internal/database"
	"github.com/sounishnath003/jobprocessor/internal/models"
)

// InitCore - helps to initialize the Core dependendies of the Jobprocessor service
// Initializrs the source database pools, result backends, queues and additional configurations
// Onestop solution to be kept running.
//
// Returns *Core, error
func InitCore(conf ConfigOpts) (*Core, error) {

	lo := slog.Default()

	// Connect to source DBs
	srcDBs := []database.Config{
		{Type: "postgres", DSN: "postgres://root:password@127.0.0.1:5432/postgres?sslmode=disable", MaxIdleConns: 10, MaxActiveConns: 100, ConnectTimeout: 10 * time.Second},
	}

	// Initializing the source pools database connections
	// to fetch data from the the databases
	srcPools, err := database.New(srcDBs)
	if err != nil {
		lo.Info("an error occured while connecting the source DBs", err)
		return nil, err
	}

	// Initializes the Result backend systems
	// Where the joboutput shall be kept, everyjob can define their
	// desired location to be kept. // whenever job completes
	// you have to fetch the data from the result backend (jobid) tables
	resultDBs := []database.Config{
		{Type: "postgres", DSN: "postgres://root:password@127.0.0.1:5432/postgres?sslmode=disable", MaxIdleConns: 10, MaxActiveConns: 100, ConnectTimeout: 10 * time.Second},
	}

	resultPools, err := database.New(resultDBs)
	if err != nil {
		lo.Info("an error occured while connecting the source DBs", err)
		return nil, err
	}

	// Initialize the result backend controller for every backend.
	backends := make(ResultBackends)
	for name, resDB := range resultPools {
		opt := database.Opt{
			DBType:         "postgres",
			ResultTable:    fmt.Sprintf("results.%s.results_table", name),
			UnloggedTables: true,
		}

		backend, err := database.NewResultSQLBackend(resDB, opt)
		if err != nil {
			return nil, fmt.Errorf("error initializing error backend: %w", err)
		}
		backends[name] = backend
	}

	// Initialize the Redis Broker backend.
	rBroker := bredis.New(bredis.Options{
		PollPeriod:   bredis.DefaultPollPeriod,
		Addrs:        conf.BrokerQueue.Addrs,
		Password:     conf.BrokerQueue.Password,
		DB:           conf.BrokerQueue.DB,
		MinIdleConns: conf.BrokerQueue.MinIdleConns,
		DialTimeout:  conf.BrokerQueue.DialTimeout,
		ReadTimeout:  conf.BrokerQueue.ReadTimeout,
		WriteTimeout: conf.BrokerQueue.WriteTimeout,
	}, lo)

	// Initialize the Redis Result State backend.
	rResult := rredis.New(rredis.Options{}, lo)

	co := &Core{
		Conf:           conf,
		sourceDBs:      srcPools,
		resultBackends: backends,
		mu:             sync.RWMutex{},
		tasks:          make(Tasks),
		lo:             lo,
		Opts: Options{
			DefaultQueue:            "test",
			DefaultGroupConcurrency: 5,
			DefaultJobTTL:           5 * time.Second,
			Broker:                  rBroker,
			Results:                 rResult,
		},
	}

	if err := co.LoadTasks([]string{"sql"}); err != nil {
		return nil, fmt.Errorf("error loading tasks: %w", err)
	}

	lo.Info("core.Conf:", slog.Any("conf", co.Conf))
	lo.Info("core.Opts:", slog.Any("opt", co.Opts))

	return co, nil
}

// Start - a fully blocking worker initializers. This must only calls after all initializations are done.
func (co *Core) Start(ctx context.Context, workerName string, workerConcurrency int) error {
	qs, err := co.initQueue()
	if err != nil {
		return err
	}

	co.q = qs
	qs.Start(ctx)

	func() {
		for {

		}
	}()

	return nil
}

// initQueue creates and returns a distributed queue system (Tasqueue) and registers
// Tasks (SQL queries) to be executed. The queue system uses a broker (eg: Kafka) and stores
// job states in a state store (eg: Redis)
func (co *Core) initQueue() (*tasqueue.Server, error) {
	var err error

	qs, err := tasqueue.NewServer(tasqueue.ServerOpts{
		Broker:  co.Opts.Broker,
		Results: co.Opts.Results,
		Logger:  co.lo.Handler(),
	})

	if err != nil {
		return nil, err
	}

	// TODO: Register every SQL tasks in the queue systems as a job function.

	return qs, nil
}

// ----------- Tasks functionalities. -------------

// LoadTasks loads SQL queries from all the *.sql files in a given directory
func (co *Core) LoadTasks(dirs []string) error {
	for _, d := range dirs {
		co.lo.Info("loading SQL queries from directory", d, "")
		tasks, err := co.loadTasks(d)
		if err != nil {
			return err
		}

		for t, q := range tasks {
			if _, ok := co.tasks[t]; ok {
				return fmt.Errorf("duplicate task %s", t)
			}

			// Add the task q into co.tasks[t]
			co.tasks[t] = q
			co.lo.Info("loaded tasks (SQL queries)", slog.Int("count", len(tasks)), "directory", d) // Updated to use slog.Int
		}
	}

	return nil
}

func (co *Core) loadTasks(dir string) (Tasks, error) {
	// Discovers .sql files.
	files, err := filepath.Glob(path.Join(dir, "*.sql"))
	if err != nil {
		return nil, fmt.Errorf("unable to read SQL files directory %s : %w", dir, err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("NO SQL files present directory %s : %w", dir, err)
	}

	// Parse all discovered SQL files.
	tasks := make(Tasks)
	for _, f := range files {
		q := goyesql.MustParseFile(f)

		for name, s := range q {
			var (
				stmt   *sql.Stmt
				srcDBs database.Pool
				resDBs ResultBackends
			)

			// Query already exists.
			if _, ok := tasks[string(name)]; ok {
				return nil, fmt.Errorf("duplicate query %s (%s)", name, f)
			}

			srcDBs = co.sourceDBs
			resDBs = co.resultBackends

			// Prepare the statement?
			typ := ""
			if _, ok := s.Tags["raw"]; ok {
				typ = "raw"
			} else {
				// Prepare the statement against all tagged DBs just to be sure.
				typ = "prepared"
				for _, db := range srcDBs {
					_, err := db.Prepare(s.Query)
					if err != nil {
						return nil, fmt.Errorf("error preparing SQL query %s: %v", name, err)
					}
				}
			}

			// Is there a queue?
			queue := co.Opts.DefaultQueue
			if v, ok := s.Tags["queue"]; ok {
				queue = strings.TrimSpace(v)
			}

			conc := co.Opts.DefaultGroupConcurrency

			co.lo.Info("loading task:", name, "type", typ, "queue", queue, "details") // Added "details" as a final value

			tasks[name] = Task{
				Name:          name,
				Queue:         queue,
				Conc:          conc,
				Stmt:          stmt,
				Raw:           s.Query,
				DBs:           srcDBs,
				ResultBackend: resDBs,
			}
		}
	}

	return tasks, nil
}

// ----------- API Handlers Dependent Jobs functionalities. -------------

// GetTasks returns the registered tasks map.
func (co *Core) GetTasks() Tasks {
	return co.tasks
}

func (co *Core) GetJobStatus(jobID string) (models.JobStatusResp, error) {

	ctx := context.Background()
	_, err := co.q.GetJob(ctx, jobID)
	if err == tasqueue.ErrNotFound {
		return models.JobStatusResp{}, fmt.Errorf("job not found")
	} else if err != nil {
		co.lo.Info("error fetching job status", "error", err)
		return models.JobStatusResp{}, err
	}

	return models.JobStatusResp{}, nil
}
