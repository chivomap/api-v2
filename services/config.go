package services

import "chivomap.com/config"

// ConfigService implements the ConfigService interface using the global config
type ConfigService struct {
	config *config.Config
}

// NewConfigService creates a new config service
func NewConfigService(cfg *config.Config) *ConfigService {
	return &ConfigService{config: cfg}
}

// NewConfigServiceFromGlobal creates a config service from the global AppConfig
func NewConfigServiceFromGlobal() *ConfigService {
	return &ConfigService{config: &config.AppConfig}
}

// GetServerPort returns the server port
func (c *ConfigService) GetServerPort() string {
	return c.config.ServerPort
}

// GetDatabaseURL returns the database URL
func (c *ConfigService) GetDatabaseURL() string {
	return c.config.DatabaseURL
}

// GetDatabaseToken returns the database token
func (c *ConfigService) GetDatabaseToken() string {
	return c.config.DatabaseToken
}

// GetBaseDir returns the base directory
func (c *ConfigService) GetBaseDir() string {
	return c.config.BaseDir
}

// GetAssetsDir returns the assets directory
func (c *ConfigService) GetAssetsDir() string {
	return c.config.AssetsDir
}