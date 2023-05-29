package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

// Holds all application db models.
type Models struct {
	Movies MovieModel
}

// Simple helper to initialize all db models with the provided db connection.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
