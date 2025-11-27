package scanner

import (
	"path/filepath"
	"strings"
)

// Filter provides file filtering capabilities for the scanner
type Filter struct {
	includeExtensions map[string]bool
	excludeExtensions map[string]bool
	excludeDirs       map[string]bool
	minFileSize       int64
	maxFileSize       int64
}

// NewFilter creates a new file filter
func NewFilter() *Filter {
	return &Filter{
		includeExtensions: make(map[string]bool),
		excludeExtensions: make(map[string]bool),
		excludeDirs:       make(map[string]bool),
		minFileSize:       0,
		maxFileSize:       0,
	}
}

// AddIncludeExtension adds file extensions to include
func (f *Filter) AddIncludeExtension(extensions ...string) {
	for _, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		f.includeExtensions[strings.ToLower(ext)] = true
	}
}

// AddExcludeExtension adds file extensions to exclude
func (f *Filter) AddExcludeExtension(extensions ...string) {
	for _, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		f.excludeExtensions[strings.ToLower(ext)] = true
	}
}

// AddExcludeDir adds directory names to exclude
func (f *Filter) AddExcludeDir(dirs ...string) {
	for _, dir := range dirs {
		f.excludeDirs[strings.ToLower(dir)] = true
	}
}

// SetSizeLimits sets minimum and maximum file size limits
func (f *Filter) SetSizeLimits(minSize, maxSize int64) {
	f.minFileSize = minSize
	f.maxFileSize = maxSize
}

// ShouldIncludeFile checks if a file should be included based on filters
func (f *Filter) ShouldIncludeFile(filePath string, fileSize int64) bool {
	// Check file extension
	if !f.isExtensionAllowed(filePath) {
		return false
	}

	// Check file size
	if !f.isSizeAllowed(fileSize) {
		return false
	}

	return true
}

// ShouldIncludeDir checks if a directory should be included based on filters
func (f *Filter) ShouldIncludeDir(dirPath string) bool {
	dirName := filepath.Base(dirPath)
	return !f.excludeDirs[strings.ToLower(dirName)]
}

// isExtensionAllowed checks if file extension is allowed
func (f *Filter) isExtensionAllowed(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	// If include list is specified, only allow those extensions
	if len(f.includeExtensions) > 0 {
		if !f.includeExtensions[ext] {
			return false
		}
	}

	// Check exclude list
	if f.excludeExtensions[ext] {
		return false
	}

	return true
}

// isSizeAllowed checks if file size is within limits
func (f *Filter) isSizeAllowed(fileSize int64) bool {
	if f.minFileSize > 0 && fileSize < f.minFileSize {
		return false
	}
	if f.maxFileSize > 0 && fileSize > f.maxFileSize {
		return false
	}
	return true
}

// GetDefaultImageFilter returns a filter configured for common image formats
func GetDefaultImageFilter() *Filter {
	filter := NewFilter()
	filter.AddIncludeExtension(
		".jpg", ".jpeg", ".png", ".webp",
		".tiff", ".tif", ".bmp", ".gif",
	)
	filter.AddExcludeDir(
		".git", ".svn", ".hg",
		"node_modules", "__pycache__",
		"thumbs", "thumbnails", ".thumbnails",
	)
	filter.SetSizeLimits(1024, 500*1024*1024) // 1KB to 500MB
	return filter
}

// GetSupportedExtensions returns list of supported image extensions
func (f *Filter) GetSupportedExtensions() []string {
	extensions := make([]string, 0, len(f.includeExtensions))
	for ext := range f.includeExtensions {
		extensions = append(extensions, ext)
	}
	return extensions
}
