package models

import (
	"database/sql"

	// MySQL database driver
	_ "github.com/go-sql-driver/mysql"
)

// Datastore defines the CRUD operations of models in the database
type Datastore interface{}

// DB represents an SQL database connection
type DB struct {
	*sql.DB
}

// NewDB initializes a new DB
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
