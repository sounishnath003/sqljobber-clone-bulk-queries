package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sounishnath003/jobprocessor/internal/models"
)

type Opt struct {
	DBType         string
	ResultTable    string
	UnloggedTables bool
}

type TableSchema struct {
	dropTable   string
	createTable string
	insertRow   string
}

type ResultDB struct {
	db     *sql.DB
	opt    Opt
	logger *log.Logger

	resultTableSchemas map[string]TableSchema
	schemaMutex        sync.RWMutex
}

type SQLDBResultSet struct {
	jobID       string
	taskName    string
	colsWritten bool
	cols        []string
	rows        [][]byte
	tx          *sql.Tx
	tbl         string

	backend *ResultDB
}

// NewResultSQLBackend returns a new ResultDB result backend instance
func NewResultSQLBackend(db *sql.DB, opt Opt) (*ResultDB, error) {
	s := ResultDB{
		db:     db,
		opt:    opt,
		logger: log.Default(),

		resultTableSchemas: make(map[string]TableSchema),
		schemaMutex:        sync.RWMutex{},
	}
	// Config. setup
	if len(opt.ResultTable) == 0 {
		s.opt.ResultTable = "results_%s"
	} else {
		s.opt.ResultTable = opt.ResultTable
	}

	return &s, nil
}

func (s *ResultDB) NewResultSet(jobID, taskName string, ttl time.Duration) (models.ResultSet, error) {
	// init through db transaction
	tx, err := s.db.Begin()
	if err != nil {
		log.Println("an error occured", err)
		return nil, err
	}

	return &SQLDBResultSet{
		jobID:    jobID,
		taskName: taskName,
		backend:  s,
		tx:       tx,
		tbl:      fmt.Sprintf(s.opt.ResultTable, jobID),
	}, nil
}
