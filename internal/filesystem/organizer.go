package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Organizer handles safe file operations with conflict resolution
type Organizer struct {
	logger *logrus.Logger
}

// NewOrganizer creates a new file organizer
func NewOrganizer() *Organizer {
	return &Organizer{
		logger: logrus.New(),
	}
}

// MoveFile safely moves a file with conflict resolution
func (o *Organizer) MoveFile(sourcePath, destDir string) (string, error) {
	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	filename := filepath.Base(sourcePath)
	destPath := filepath.Join(destDir, filename)

	// Resolve naming conflicts
	destPath = o.resolveConflict(destPath)

	// Perform the move operation
	if err := os.Rename(sourcePath, destPath); err != nil {
		return "", fmt.Errorf("failed to move file: %w", err)
	}

	o.logger.Debugf("Moved file: %s -> %s", sourcePath, destPath)
	return destPath, nil
}

// CopyFile safely copies a file with conflict resolution
func (o *Organizer) CopyFile(sourcePath, destDir string) (string, error) {
	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	filename := filepath.Base(sourcePath)
	destPath := filepath.Join(destDir, filename)

	// Resolve naming conflicts
	destPath = o.resolveConflict(destPath)

	// Read source file
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write destination file: %w", err)
	}

	o.logger.Debugf("Copied file: %s -> %s", sourcePath, destPath)
	return destPath, nil
}

// DeleteFile safely deletes a file with backup option
func (o *Organizer) DeleteFile(filePath string, backupDir string) error {
	// Create backup if requested
	if backupDir != "" {
		backupPath, err := o.CopyFile(filePath, backupDir)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		o.logger.Debugf("Created backup: %s", backupPath)
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	o.logger.Debugf("Deleted file: %s", filePath)
	return nil
}

// resolveConflict handles filename conflicts by appending counters or timestamps
func (o *Organizer) resolveConflict(originalPath string) string {
	if !o.fileExists(originalPath) {
		return originalPath
	}

	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Try counter-based resolution
	for i := 1; i < 1000; i++ {
		newName := fmt.Sprintf("%s_%d%s", name, i, ext)
		newPath := filepath.Join(dir, newName)
		if !o.fileExists(newPath) {
			return newPath
		}
	}

	// Fallback to timestamp-based resolution
	timestamp := time.Now().Format("20060102_150405")
	newName := fmt.Sprintf("%s_%s%s", name, timestamp, ext)
	return filepath.Join(dir, newName)
}

// fileExists checks if a file exists
func (o *Organizer) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetFileInfo returns detailed information about a file
func (o *Organizer) GetFileInfo(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return info, nil
}

// CreateDirectory creates a directory with proper permissions
func (o *Organizer) CreateDirectory(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// CleanEmptyDirectories removes empty directories recursively
func (o *Organizer) CleanEmptyDirectories(rootPath string) ([]string, error) {
	var removedDirs []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		// Skip the root directory
		if path == rootPath {
			return nil
		}

		// Check if directory is empty
		isEmpty, err := o.isDirectoryEmpty(path)
		if err != nil {
			return err
		}

		if isEmpty {
			if err := os.Remove(path); err != nil {
				return err
			}
			removedDirs = append(removedDirs, path)
			o.logger.Debugf("Removed empty directory: %s", path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to clean empty directories: %w", err)
	}

	return removedDirs, nil
}

// isDirectoryEmpty checks if a directory is empty
func (o *Organizer) isDirectoryEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}
