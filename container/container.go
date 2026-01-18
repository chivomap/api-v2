package container

import (
	"database/sql"
	"fmt"

	"chivomap.com/cache"
	"chivomap.com/interfaces"
	"chivomap.com/services"
)

// Container holds all application dependencies
type Container struct {
	Config      interfaces.ConfigService
	DB          interfaces.DatabaseService
	CensoDB     interfaces.DatabaseService
	Logger      interfaces.Logger
	StaticCache interfaces.StaticCacheService
}

// NewContainer creates a new dependency injection container
func NewContainer(config interfaces.ConfigService, db, censoDB *sql.DB) (*Container, error) {
	if config == nil {
		return nil, fmt.Errorf("config service is required")
	}
	if db == nil {
		return nil, fmt.Errorf("main database connection is required")
	}

	// Wrap sql.DB connections with our interface
	dbService := services.NewDatabaseService(db)
	var censoDBService interfaces.DatabaseService
	if censoDB != nil {
		censoDBService = services.NewDatabaseService(censoDB)
	}

	// Create static cache service
	staticCache := cache.NewStaticFileCache(config.GetAssetsDir())

	// Create logger service
	logger := services.NewLogger()

	return &Container{
		Config:      config,
		DB:          dbService,
		CensoDB:     censoDBService,
		Logger:      logger,
		StaticCache: staticCache,
	}, nil
}

// Close properly closes all resources
func (c *Container) Close() error {
	var errs []error

	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error closing main database: %w", err))
		}
	}

	if c.CensoDB != nil {
		if err := c.CensoDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error closing censo database: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing container: %v", errs)
	}

	return nil
}