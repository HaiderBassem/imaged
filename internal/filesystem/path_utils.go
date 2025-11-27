package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathUtils provides utility functions for path manipulation
type PathUtils struct{}

// NewPathUtils creates a new path utilities instance
func NewPathUtils() *PathUtils {
	return &PathUtils{}
}

// NormalizePath normalizes a file path for consistent comparison
func (pu *PathUtils) NormalizePath(path string) string {
	// Clean the path
	cleanPath := filepath.Clean(path)

	// Convert to absolute path if relative
	if !filepath.IsAbs(cleanPath) {
		absPath, err := filepath.Abs(cleanPath)
		if err == nil {
			cleanPath = absPath
		}
	}

	// Normalize separators (Windows support)
	cleanPath = filepath.ToSlash(cleanPath)

	return cleanPath
}

// IsSubpath checks if a path is a subpath of another
func (pu *PathUtils) IsSubpath(parent, child string) bool {
	parent = pu.NormalizePath(parent)
	child = pu.NormalizePath(child)

	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, "..") && rel != "."
}

// GetRelativePath returns the relative path from base to target
func (pu *PathUtils) GetRelativePath(base, target string) (string, error) {
	base = pu.NormalizePath(base)
	target = pu.NormalizePath(target)

	return filepath.Rel(base, target)
}

// SafeJoin safely joins path elements while preventing directory traversal
func (pu *PathUtils) SafeJoin(base string, elems ...string) (string, error) {
	fullPath := filepath.Join(append([]string{base}, elems...)...)

	// Check if the resulting path is still within the base directory
	if !pu.IsSubpath(base, fullPath) {
		return "", fmt.Errorf("path traversal attempt detected: %s", fullPath)
	}

	return fullPath, nil
}

// GetFileExtension returns the lowercase file extension
func (pu *PathUtils) GetFileExtension(path string) string {
	ext := filepath.Ext(path)
	return strings.ToLower(ext)
}

// ChangeExtension changes the file extension
func (pu *PathUtils) ChangeExtension(path, newExt string) string {
	if !strings.HasPrefix(newExt, ".") {
		newExt = "." + newExt
	}

	base := strings.TrimSuffix(path, filepath.Ext(path))
	return base + newExt
}

// IsHiddenFile checks if a file is hidden (starts with .)
func (pu *PathUtils) IsHiddenFile(path string) bool {
	filename := filepath.Base(path)
	return strings.HasPrefix(filename, ".")
}

// GetDirectorySize calculates the total size of a directory
func (pu *PathUtils) GetDirectorySize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}
