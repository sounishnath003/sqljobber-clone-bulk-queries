package core

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sounishnath003/jobprocessor/internal/database"
)

// InitCore - helps to initialize the Core dependendies of the Jobprocessor service
// Initializrs the source database pools, result backends, queues and additional configurations
// Onestop solution to be kept running.
//
// Returns *Core, error
func InitCore(conf ConfigOpts) (*Core, error) {

	// Connect to source DBs
	srcDBs := []database.Config{
		{Type: "postgres", DSN: "postgres://root:password@127.0.0.1:5432/postgres?sslmode=disable", MaxIdleConns: 10, MaxActiveConns: 100, ConnectTimeout: 10 * time.Second},
	}

	// Initializing the source pools database connections
	// to fetch data from the the databases
	srcPools, err := database.New(srcDBs)
	if err != nil {
		log.Println("an error occured while connecting the source DBs", err)
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
		log.Println("an error occured while connecting the source DBs", err)
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

	co := &Core{
		Conf:           conf,
		sourceDBs:      srcPools,
		resultBackends: backends,
		mu:             sync.RWMutex{},
		Opts: Options{
			DefaultQueue:            "default-queue",
			DefaultGroupConcurrency: 10,
			DefaultJobTTL:           10 * time.Second,
		},
	}

	log.Printf("core.Conf: %v\n", co.Conf)
	log.Printf("core.Opts: %v\n", co.Opts)

	return co, nil
}
