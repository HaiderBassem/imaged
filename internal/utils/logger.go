package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

// Logger provides enhanced logging capabilities
type Logger struct {
	*logrus.Logger
	config LogConfig
}

// LogConfig defines logging configuration
type LogConfig struct {
	Level      string
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

// NewLogger creates a new configured logger
func NewLogger(config LogConfig) (*Logger, error) {
	logger := &Logger{
		Logger: logrus.New(),
		config: config,
	}

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set formatter
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := filepath.Base(f.File)
			return "", fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})

	// Setup file logging if specified
	if config.FilePath != "" {
		if err := logger.setupFileOutput(); err != nil {
			return nil, err
		}
	}

	return logger, nil
}

// setupFileOutput configures file-based logging
func (l *Logger) setupFileOutput() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(l.config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(l.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.SetOutput(file)
	return nil
}

// WithContext adds contextual information to the logger
func (l *Logger) WithContext(fields map[string]interface{}) *logrus.Entry {
	return l.WithFields(logrus.Fields(fields))
}

// LogOperation logs the start and completion of an operation
func (l *Logger) LogOperation(operation string, fn func() error) error {
	l.Infof("Starting operation: %s", operation)

	err := fn()

	if err != nil {
		l.Errorf("Operation failed: %s - %v", operation, err)
	} else {
		l.Infof("Operation completed: %s", operation)
	}

	return err
}

// LogProgress logs progress information
func (l *Logger) LogProgress(operation string, current, total int) {
	if current%100 == 0 || current == total { // Log every 100 items or when complete
		percentage := float64(current) / float64(total) * 100
		l.Infof("%s progress: %d/%d (%.1f%%)", operation, current, total, percentage)
	}
}

// LogPerformance logs performance metrics
func (l *Logger) LogPerformance(operation string, duration int64, itemsProcessed int) {
	itemsPerSecond := float64(itemsProcessed) / (float64(duration) / 1e9)
	l.Infof("Performance: %s - %d items in %v (%.1f items/sec)",
		operation, itemsProcessed, duration, itemsPerSecond)
}

// GetDefaultConfig returns default logging configuration
func GetDefaultConfig() LogConfig {
	return LogConfig{
		Level:      "info",
		FilePath:   "",
		MaxSize:    100, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}
}

// CreateModuleLogger creates a logger for a specific module
func CreateModuleLogger(module string, config LogConfig) (*Logger, error) {
	logger, err := NewLogger(config)
	if err != nil {
		return nil, err
	}

	// Add module field to all log entries
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return logger, nil
}
