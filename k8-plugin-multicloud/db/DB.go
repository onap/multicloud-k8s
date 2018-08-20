package db

import (
	pkgerrors "github.com/pkg/errors"
)

// DBconn interface used to talk a concrete Database connection
var DBconn DatabaseConnection

// DatabaseConnection is an interface for accessing a database
type DatabaseConnection interface {
	InitializeDatabase() error
	CheckDatabase() error
	CreateEntry(string, string) error
	ReadEntry(string) (string, bool, error)
	DeleteEntry(string) error
	ReadAll(string) ([]string, error)
}

// CreateDBClient creates the DB client
var CreateDBClient = func(dbType string) error {
	if dbType == "consul" {
		DBconn = &ConsulDB{}
		return nil
	}

	return pkgerrors.New("No suitable DB found")
}
