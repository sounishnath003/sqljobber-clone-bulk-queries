package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sounishnath003/jobprocessor/internal/models"
)

const (
	dbTypePostgres = "postgres"
	dbTypeMysql    = "mysql"
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

// insertSchema contains the generated SQL for creating tables
// and inserting rows.
type insertSchema struct {
	dropTable   string
	createTable string
	insertRow   string
}

type ResultDB struct {
	db     *sql.DB
	opt    Opt
	logger *log.Logger

	resultTableSchemas map[string]insertSchema
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

		resultTableSchemas: make(map[string]insertSchema),
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

// IsColTypesRegistered checks whether the column types for a particular taskname's
// structure is registered in the backend.
func (w *SQLDBResultSet) IsColTypesRegistered() bool {
	w.backend.schemaMutex.RLock()
	_, ok := w.backend.resultTableSchemas[w.taskName]
	w.backend.schemaMutex.RUnlock()

	return ok
}

// RegisterColTypes registers the column types of a particular taskName's result set.
// Internally, it translates sql types into the simpler sqlDB (SQLite 3) types,
// creates a CREATE TABLE() schema for the results table with the structure of the
// particular taskName, and caches it be used for every subsequent result db creation
// and population. This should only be called once for each kind of taskName.
func (w *SQLDBResultSet) RegisterColTypes(cols []string, colTypes []*sql.ColumnType) error {
	// Check for registration?
	if w.IsColTypesRegistered() {
		return errors.New("column types are already registered!")
	}

	w.cols = make([]string, len(cols))
	copy(w.cols, cols)

	// Create the INSERT statement
	// INSERT INTO xxx (col1, col2...) VALUES.
	var (
		colNameHeader = make([]string, len(cols))
		colValHeader  = make([]string, len(cols))
	)

	for i := range w.cols {
		colNameHeader[i] = fmt.Sprintf(`"%s"`, w.cols[i])
		colValHeader[i] = "?" // As result backend POSTGRES is only used
	}

	ins := fmt.Sprintf(`INSERT INTO "%%s" (%s)`, strings.Join(colNameHeader, ","))
	ins += fmt.Sprintf(`VALUES (%s)`, strings.Join(colValHeader, ","))

	w.backend.schemaMutex.RLock()
	w.backend.resultTableSchemas[w.taskName] = w.createTableSchema(cols, colTypes)
	w.backend.schemaMutex.RUnlock()

	return nil
}

// func (w *SQLDBResultSet) IsColTypesRegistered() bool {}

// WriteCols([]string) error
// WriteRow([]interface{}) error
// Flush() error
// Close() error

// createTableSchema takes an SQL query results, gets its column names and types,
// and generates a sqlDB CREATE TABLE() schema for the results.
func (s *SQLDBResultSet) createTableSchema(cols []string, colTypes []*sql.ColumnType) insertSchema {
	var (
		colNameHolder = make([]string, len(cols))
		colValHolder  = make([]string, len(cols))
	)

	for i := range cols {
		colNameHolder[i] = fmt.Sprintf(`"%s"`, cols[i])

		// This will be filled by the driver.
		if s.backend.opt.DBType == dbTypePostgres {
			// Postgres placeholders are $1, $2 ...
			colValHolder[i] = fmt.Sprintf("$%d", i+1)
		} else {
			colValHolder[i] = "?"
		}
	}

	var (
		fields   = make([]string, len(cols))
		typ      = ""
		unlogged = ""
	)

	for i := 0; i < len(cols); i++ {
		typ = colTypes[i].DatabaseTypeName()
		switch colTypes[i].DatabaseTypeName() {
		case "INT2", "INT4", "INT8", // Postgres
			"TINYINT", "SMALLINT", "INT", "MEDIUMINT", "BIGINT": // MySQL
			typ = "BIGINT"
		case "FLOAT4", "FLOAT8", // Postgres
			"DECIMAL", "FLOAT", "DOUBLE", "NUMERIC": // MySQL
			typ = "DECIMAL"
		case "TIMESTAMP", // Postgres, MySQL
			"DATETIME": // MySQL
			typ = "TIMESTAMP"
		case "DATE": // Postgres, MySQL
			typ = "DATE"
		case "BOOLEAN": // Postgres, MySQL
			typ = "BOOLEAN"
		case "JSON", "JSONB": // Postgres
			if s.backend.opt.DBType != dbTypePostgres {
				typ = "TEXT"
			}
		// _INT4, _INT8, _TEXT represent array types in Postgres
		case "_INT4": // Postgres
			typ = "_INT4"
		case "_INT8": // Postgres
			typ = "_INT8"
		case "_TEXT": // Postgres
			typ = "_TEXT"
		default:
			typ = "TEXT"
		}

		if nullable, ok := colTypes[i].Nullable(); ok && !nullable {
			typ += " NOT NULL"
		}

		fields[i] = fmt.Sprintf(`"%s" %s`, cols[i], typ)
	}

	// If the DB is Postgres, optionally create an "unlogged" table that disables
	// WAL, improving performance of throw-away cache tables.
	// https://www.postgresql.org/docs/current/sql-createtable.html
	if s.backend.opt.DBType == dbTypePostgres && s.backend.opt.UnloggedTables {
		unlogged = "UNLOGGED"
	}

	return insertSchema{
		dropTable:   `DROP TABLE IF EXISTS "%s";`,
		createTable: fmt.Sprintf(`CREATE %s TABLE IF NOT EXISTS "%%s" (%s);`, unlogged, strings.Join(fields, ",")),
		insertRow: fmt.Sprintf(`INSERT INTO "%%s" (%s) VALUES (%s)`, strings.Join(colNameHolder, ","),
			strings.Join(colValHolder, ",")),
	}
}

// WriteCols writes the column (headers) of a result set to the backend.
// Internally, it creates a sqlDB database and creates a results table
// based on the schema RegisterColTypes() would've created and cached.
// This should only be called once on a ResultWriter instance.
func (w *SQLDBResultSet) WriteCols(cols []string) error {
	if w.colsWritten {
		return fmt.Errorf("columns for '%s' are already written", w.taskName)
	}

	w.backend.schemaMutex.RLock()
	rSchema, ok := w.backend.resultTableSchemas[w.taskName]
	w.backend.schemaMutex.RUnlock()

	if !ok {
		return fmt.Errorf("column types for '%s' have not been registered", w.taskName)
	}

	// Create the results table.
	tx, err := w.backend.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(fmt.Sprintf(rSchema.dropTable, w.tbl)); err != nil {
		return err
	}

	if _, err := tx.Exec(fmt.Sprintf(rSchema.createTable, w.tbl)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return err
}

// WriteRow writes an individual row from a result set to the backend.
// Internally, it INSERT()s the given row into the sqlDB results table.
func (w *SQLDBResultSet) WriteRow(row []interface{}) error {
	w.backend.schemaMutex.RLock()
	rSchema, ok := w.backend.resultTableSchemas[w.taskName]
	w.backend.schemaMutex.RUnlock()

	if !ok {
		return fmt.Errorf("column types for '%s' have not been registered", w.taskName)
	}

	_, err := w.tx.Exec(fmt.Sprintf(rSchema.insertRow, w.tbl), row...)

	return err
}

// Flush flushes the rows written into the sqlDB pipe.
func (w *SQLDBResultSet) Flush() error {
	err := w.tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Close closes the active sqlDB connection.
func (w *SQLDBResultSet) Close() error {
	if w.tx != nil {
		return w.tx.Rollback()
	}

	return nil
}
