package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	uuid "github.com/gofrs/uuid/v5"
	"github.com/kalbhor/tasqueue/v2"
	bredis "github.com/kalbhor/tasqueue/v2/brokers/redis"
	rredis "github.com/kalbhor/tasqueue/v2/results/redis"
	"github.com/vmihailenco/msgpack"

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
		jobCtx:         make(map[string]context.CancelFunc),
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
	for name := range co.tasks {
		query := co.tasks[name]
		err := qs.RegisterTask(string(name), func(b []byte, jctx tasqueue.JobCtx) error {
			if _, err := qs.GetJob(jctx, jctx.Meta.ID); err != nil {
				return err
			}

			var args taskMeta
			if err := msgpack.Unmarshal(b, &args); err != nil {
				return fmt.Errorf("could not unmarshal args: %w", err)
			}

			count, err := co.execJob(jctx.Meta.ID, name, args.DB, time.Duration(args.TTL)*time.Second, args.Args, query)
			if err != nil {
				return fmt.Errorf("could not execute the job: %w", err)
			}

			return jctx.Save([]byte(strconv.Itoa(int(count))))
		}, tasqueue.TaskOpts{
			Concurrency: uint32(query.Conc),
			Queue:       query.Queue,
		})

		if err != nil {
			return nil, err
		}
	}

	return qs, nil
}

// execJob executes an SQL statement job and inserts the results into the result backend.
func (co *Core) execJob(jobID, taskName, dbName string, ttl time.Duration, args []interface{}, task Task) (int64, error) {
	// If the job is deleted, stop.
	if _, err := co.GetJobStatus(jobID); err != nil {
		return 0, errors.New("job was cancelled")
	}

	// Execute the query.
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()

		co.mu.Lock()
		delete(co.jobCtx, jobID)
		co.mu.Unlock()
	}()

	co.mu.Lock()
	co.jobCtx[jobID] = cancel
	co.mu.Unlock()

	var (
		rows *sql.Rows
		db   *sql.DB
		err  error
	)

	if task.Stmt != nil {
		// Prepared query.
		rows, err := task.Stmt.QueryContext(ctx, args...)
		if err != nil {
			return 0, fmt.Errorf("query preparation failed: %w", err)
		}
	} else {
		if dbName != "" {
			// Specific DB.
			d, err := task.DBs.Get(dbName)
			if err != nil {
				return 0, fmt.Errorf("task execution failed : %w", err)
			}
			db = d
		} else {
			// Random DB as NO DB provided to run the query?
			return 0, fmt.Errorf("database is not specified %w", err)
		}
	}

	rows, err := db.QueryContext(ctx, task.Raw, args...)
	if err != nil {
		if err == context.Canceled {
			return 0, fmt.Errorf("job was cancelled : %w", err)
		}

		return 0, fmt.Errorf("task execution %s failed : %w", taskName, err)
	}

	defer rows.Close()

	// Write the results into result backend.
	return co.writeResults(jobID, task, ttl, rows)
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

// makeJob creates and returns a tasqueue (tasqueue.Job{}) from the given job params.
func (co *Core) makeJob(j models.JobReq, taskName string) (tasqueue.Job, error) {
	// Find the taskName is present or not.
	task, ok := co.tasks[taskName]
	if !ok {
		return tasqueue.Job{}, fmt.Errorf("uncognized task: %s", taskName)
	}

	// Check if a JOB with same ID already running.
	msg, err := co.q.GetJob(context.Background(), j.JobID)
	if err != nil {
		return tasqueue.Job{}, fmt.Errorf("error fetching job status (id='%s')", j.JobID)
	}
	if msg.Status == tasqueue.StatusProcessing || msg.Status == tasqueue.StatusRetrying {
		return tasqueue.Job{}, fmt.Errorf("job '%s' is already running", j.JobID)
	}

	// If there's no job_id, we generate one. This is because
	// the ID Machinery generates is not made available inside the
	// actual task. So, we generate the ID and pass it as an argument
	// to the task itself.
	if j.JobID == "" {
		uid, err := uuid.NewV4()
		if err != nil {
			return tasqueue.Job{}, fmt.Errorf("error generating uuid: %v", err)
		}
		j.JobID = fmt.Sprintf("job_%v", uid.String())
	}

	ttl := co.Opts.DefaultJobTTL
	if j.TTL > 0 {
		ttl = time.Duration(j.TTL) * time.Second
	}

	var eta time.Time
	if j.ETA != "" {
		e, err := time.Parse("2006-01-02 15:04:05", j.ETA)
		if err != nil {
			return tasqueue.Job{}, fmt.Errorf("error parsing ETA: %v", err)
		}

		eta = e
	}

	// If there's no queue in the request, use the one attached to the task,
	// if there's any (Machinery will use the default queue otherwise).
	if j.Queue == "" {
		j.Queue = task.Queue
	}

	var args = make([]interface{}, len(j.Args))
	for i := range j.Args {
		args[i] = j.Args[i]
	}

	b, err := msgpack.Marshal(taskMeta{
		Args: args,
		DB:   j.DB,
		TTL:  int(ttl),
	})
	if err != nil {
		return tasqueue.Job{}, err
	}

	return tasqueue.NewJob(taskName, b, tasqueue.JobOpts{
		ID:         j.JobID,
		Queue:      j.Queue,
		MaxRetries: uint32(j.Retries),
		ETA:        eta,
	})
}

type taskMeta struct {
	Args []interface{} `json:"args"`
	DB   string        `json:"db"`
	TTL  int           `json:"ttl"`
}

// ----------- API Handlers Dependent Jobs functionalities. -------------

// GetTasks returns the registered tasks map.
func (co *Core) GetTasks() Tasks {
	return co.tasks
}

func (co *Core) NewJob(j models.JobReq, taskName string) (models.JobResp, error) {
	sig, err := co.makeJob(j, taskName)
	if err != nil {
		return models.JobResp{}, err
	}

	// Create job.
	uuid, err := co.q.Enqueue(context.Background(), sig)
	if err != nil {
		return models.JobResp{}, err
	}

	return models.JobResp{
		JobID:    uuid,
		TaskName: sig.Task,
		Queue:    sig.Opts.Queue,
		Retries:  int(sig.Opts.MaxRetries),
		ETA:      &sig.Opts.ETA,
	}, nil
}

func (co *Core) GetJobStatus(jobID string) (models.JobStatusResp, error) {

	ctx := context.Background()
	out, err := co.q.GetJob(ctx, jobID)
	if err == tasqueue.ErrNotFound {
		return models.JobStatusResp{}, fmt.Errorf("job not found")
	} else if err != nil {
		co.lo.Info("error fetching job status", "error", err)
		return models.JobStatusResp{}, err
	}

	// Try fetcing the job's result and ignore in case.
	// The result is not found (job could be in processing)
	// Read tasqueue documentation.
	res, err := co.q.GetResult(ctx, jobID)
	if err != nil && !errors.Is(err, tasqueue.ErrNotFound) {
		co.lo.Error("error fetching jobID results", "error", err)
		return models.JobStatusResp{}, err
	}

	var rowCount int
	if string(res) != "" {
		rowCount, err = strconv.Atoi(string(res))
		if err != nil {
			co.lo.Error("error converting row count to int", "error", err)
			return models.JobStatusResp{}, err
		}
	}

	state, err := getState(out.Status)
	if err != nil {
		co.lo.Error("error fetching job status mapping", "error", err)
		return models.JobStatusResp{}, err
	}

	return models.JobStatusResp{
		JobID: out.ID,
		State: state,
		Error: out.PrevErr,
		Count: rowCount,
	}, nil
}

// GetPendingJobs returns jobs pending execution.
func (co *Core) GetPendingJobs(queue string) ([]tasqueue.JobMessage, error) {
	out, err := co.q.GetPending(context.Background(), queue)
	if err != nil {
		co.lo.Error("error fetching pending jobs", "queue", queue, "error", err)
		return nil, err
	}

	jLen := len(out)
	for i := 0; i < jLen; i++ {
		out[i], out[jLen-i-1] = out[jLen-i-1], out[i]
	}

	return out, nil
}

// getState helps keep compatibility between different status conventions
// of the job servers (tasqueue/machinery).
func getState(st string) (string, error) {
	switch st {
	case tasqueue.StatusStarted:
		return models.StatusPending, nil
	case tasqueue.StatusProcessing:
		return models.StatusStarted, nil
	case tasqueue.StatusFailed:
		return models.StatusFailure, nil
	case tasqueue.StatusDone:
		return models.StatusSuccess, nil
	case tasqueue.StatusRetrying:
		return models.StatusRetry, nil
	}

	return "", fmt.Errorf("invalid status not found in mapping")
}
