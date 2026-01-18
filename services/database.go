package services

import (
	"context"
	"database/sql"
	"fmt"
)

// DatabaseService wraps sql.DB to implement our interface
type DatabaseService struct {
	db *sql.DB
}

// NewDatabaseService creates a new database service wrapper
func NewDatabaseService(db *sql.DB) *DatabaseService {
	return &DatabaseService{db: db}
}

// QueryContext executes a query that returns rows
func (d *DatabaseService) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	return rows, nil
}

// QueryRowContext executes a query that returns at most one row
func (d *DatabaseService) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// ExecContext executes a query that doesn't return rows
func (d *DatabaseService) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	result, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing statement: %w", err)
	}
	return result, nil
}

// Ping verifies the database connection
func (d *DatabaseService) Ping() error {
	if err := d.db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// Close closes the database connection
func (d *DatabaseService) Close() error {
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("error closing database connection: %w", err)
	}
	return nil
}