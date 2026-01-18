package interfaces

import (
	"context"
	"database/sql"

	"chivomap.com/types"
)

// ConfigService provides access to application configuration
type ConfigService interface {
	GetServerPort() string
	GetDatabaseURL() string
	GetDatabaseToken() string
	GetBaseDir() string
	GetAssetsDir() string
}

// DatabaseService provides database operations
type DatabaseService interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Ping() error
	Close() error
}

// StaticCacheService provides access to cached static files
type StaticCacheService interface {
	GetGeoData() (*types.GeoFeatureCollection, error)
	LoadTopoJSON() (*types.TopoJSON, error)
	GetCacheStats() map[string]interface{}
}

// Logger provides logging functionality
type Logger interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
}