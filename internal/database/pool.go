package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Pool represents a map of *sql.DB connections.
type Pool map[string]*sql.DB

type Config struct {
	Type           string
	DSN            string
	Unlogged       bool
	MaxIdleConns   int
	MaxActiveConns int
	ConnectTimeout time.Duration
}

// New function creates and returns a pool of database connections
// It takes a slice of Configs as input and returns a Pool and an error
func New(srcDBs []Config) (Pool, error) {
	out := make(Pool, len(srcDBs))

	for _, sdb := range srcDBs {
		db, err := NewConn(sdb)
		if err != nil {
			return nil, err
		}

		out[sdb.Type] = db
	}

	return out, nil
}

// NewConn creates and returns a database connection DB instance
func NewConn(cfg Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.Type, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("error connecting to db=%w", err)
	}

	db.SetConnMaxIdleTime(cfg.ConnectTimeout * time.Second)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxActiveConns)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database is not healthy. %w", err)
	}

	return db, nil
}

// Get returns an *sql.DB instance from the database Pool
func (p Pool) Get(dbName string) (*sql.DB, error) {
	db, ok := p[dbName]
	if !ok {
		return nil, fmt.Errorf("could not find the database name : %s", dbName)
	}
	return db, nil
}
