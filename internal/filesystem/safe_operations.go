package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SafeOperations provides safe file operations with error handling
type SafeOperations struct {
	logger *logrus.Logger
}

// NewSafeOperations creates a new safe operations instance
func NewSafeOperations() *SafeOperations {
	return &SafeOperations{
		logger: logrus.New(),
	}
}

// SafeMove safely moves a file with backup and verification
func (so *SafeOperations) SafeMove(source, destination string) error {
	// Verify source exists
	if err := so.verifyFileExists(source); err != nil {
		return fmt.Errorf("source verification failed: %w", err)
	}

	// Create destination directory
	destDir := filepath.Dir(destination)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Check if destination exists
	if so.fileExists(destination) {
		// Create backup of existing file
		backupPath := so.generateBackupPath(destination)
		if err := so.copyFile(destination, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		so.logger.Debugf("Created backup: %s", backupPath)
	}

	// Perform the move
	if err := os.Rename(source, destination); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	// Verify move was successful
	if !so.fileExists(destination) {
		return fmt.Errorf("move verification failed: destination file not found")
	}

	so.logger.Infof("Successfully moved: %s -> %s", source, destination)
	return nil
}

// SafeCopy safely copies a file with verification
func (so *SafeOperations) SafeCopy(source, destination string) error {
	// Verify source exists
	if err := so.verifyFileExists(source); err != nil {
		return fmt.Errorf("source verification failed: %w", err)
	}

	// Create destination directory
	destDir := filepath.Dir(destination)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Perform the copy
	if err := so.copyFile(source, destination); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Verify copy was successful
	if !so.fileExists(destination) {
		return fmt.Errorf("copy verification failed: destination file not found")
	}

	so.logger.Debugf("Successfully copied: %s -> %s", source, destination)
	return nil
}

// SafeDelete safely deletes a file with optional backup
func (so *SafeOperations) SafeDelete(path string, backupDir string) error {
	// Verify file exists
	if err := so.verifyFileExists(path); err != nil {
		return fmt.Errorf("file verification failed: %w", err)
	}

	// Create backup if requested
	if backupDir != "" {
		backupPath := filepath.Join(backupDir, filepath.Base(path))
		if err := so.SafeCopy(path, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		so.logger.Debugf("Created backup: %s", backupPath)
	}

	// Perform deletion
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Verify deletion
	if so.fileExists(path) {
		return fmt.Errorf("deletion verification failed: file still exists")
	}

	so.logger.Infof("Successfully deleted: %s", path)
	return nil
}

// verifyFileExists verifies that a file exists and is accessible
func (so *SafeOperations) verifyFileExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	return nil
}

// fileExists checks if a file exists
func (so *SafeOperations) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// copyFile performs a file copy operation
func (so *SafeOperations) copyFile(source, destination string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	return os.WriteFile(destination, data, 0644)
}

// generateBackupPath generates a backup path with timestamp
func (so *SafeOperations) generateBackupPath(originalPath string) string {
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("%s_%s%s", name, timestamp, ext)

	return filepath.Join(dir, backupName)
}
