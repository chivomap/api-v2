package services

import "chivomap.com/utils"

// LoggerService implements the Logger interface using utils
type LoggerService struct{}

// NewLogger creates a new logger service
func NewLogger() *LoggerService {
	return &LoggerService{}
}

// Info logs an info message
func (l *LoggerService) Info(format string, args ...interface{}) {
	utils.Info(format, args...)
}

// Error logs an error message
func (l *LoggerService) Error(format string, args ...interface{}) {
	utils.Error(format, args...)
}

// Fatal logs a fatal message and exits
func (l *LoggerService) Fatal(format string, args ...interface{}) {
	utils.Fatal(format, args...)
}