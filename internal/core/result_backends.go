package core

import (
	"github.com/sounishnath003/jobprocessor/internal/models"
)

// ResultBackends represents a map of result writting backends (sql.DBs).
type ResultBackends map[string]models.ResultBackend

// Get returns an *sql.DB from the DBs map by name.
func (r ResultBackends) Get(name string) models.ResultBackend {
	return r[name]
}

// GetNames returns the list of available DB names.
func (r ResultBackends) GetNames() []string {
	var names []string
	for n := range r {
		names = append(names, n)
	}

	return names
}
