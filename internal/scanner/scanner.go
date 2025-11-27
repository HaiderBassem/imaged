package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// Scanner handles recursive directory scanning and image file discovery
type Scanner struct {
	config Config
	logger *logrus.Logger
}

// Config defines scanner behavior and supported formats
type Config struct {
	NumWorkers       int
	SupportedFormats []string
	ExcludeDirs      []string
	MaxFileSize      int64
	FollowSymlinks   bool
}

// DefaultConfig returns sensible default scanner configuration
func DefaultConfig() Config {
	return Config{
		NumWorkers:       4,
		SupportedFormats: []string{".jpg", ".jpeg", ".png", ".webp", ".tiff", ".bmp", ".gif"},
		ExcludeDirs:      []string{".git", ".svn", ".hg", "node_modules", "__pycache__"},
		MaxFileSize:      500 * 1024 * 1024, // 500MB
		FollowSymlinks:   false,
	}
}

// NewScanner creates a new directory scanner with the specified configuration
func NewScanner(cfg Config) *Scanner {
	if cfg.NumWorkers <= 0 {
		cfg.NumWorkers = 4
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Scanner{
		config: cfg,
		logger: logger,
	}
}

// ScanResult represents the outcome of a scanning operation
type ScanResult struct {
	ImagePaths []string
	Errors     []error
	TotalFiles int
}

// scanDirectory is a compatibility wrapper for worker pool
func (s *Scanner) scanDirectory(path string) ([]string, error) {
	var files []string
	var mu sync.Mutex
	err := s.processDirectory(path, &files, &mu)
	return files, err
}

// ScanFolder recursively scans a directory for image files
func (s *Scanner) ScanFolder(ctx context.Context, rootPath string) ([]string, error) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	s.logger.Infof("Starting scan of directory: %s", absPath)

	// Verify the path exists and is a directory
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}
	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	var imagePaths []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create worker pool for parallel directory processing
	jobs := make(chan string, s.config.NumWorkers*2)
	errors := make(chan error, s.config.NumWorkers*2)

	// Start worker goroutines
	for i := 0; i < s.config.NumWorkers; i++ {
		wg.Add(1)
		go s.worker(ctx, i, jobs, &imagePaths, &mu, &wg, errors)
	}

	// Start directory walker
	go s.walkDirectories(ctx, absPath, jobs, errors)

	// Wait for completion and collect errors
	go func() {
		wg.Wait()
		close(errors)
	}()

	// Process errors
	var scanErrors []error
	for err := range errors {
		if err != nil {
			scanErrors = append(scanErrors, err)
			s.logger.Warnf("Scan error: %v", err)
		}
	}

	s.logger.Infof("Scan completed. Found %d images, %d errors", len(imagePaths), len(scanErrors))

	if len(scanErrors) > 0 && len(imagePaths) == 0 {
		return nil, fmt.Errorf("scan failed with %d errors: %v", len(scanErrors), scanErrors[0])
	}

	return imagePaths, nil
}

// worker processes directories from the jobs channel
func (s *Scanner) worker(ctx context.Context, id int, jobs <-chan string, imagePaths *[]string, mu *sync.Mutex, wg *sync.WaitGroup, errors chan<- error) {
	defer wg.Done()

	for dir := range jobs {
		select {
		case <-ctx.Done():
			s.logger.Debugf("Worker %d stopping due to context cancellation", id)
			return
		default:
			if err := s.processDirectory(dir, imagePaths, mu); err != nil {
				errors <- fmt.Errorf("worker %d: %w", id, err)
			}
		}
	}
}

// walkDirectories recursively walks the directory tree and sends directories to workers
func (s *Scanner) walkDirectories(ctx context.Context, root string, jobs chan<- string, errors chan<- error) {
	defer close(jobs)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errors <- fmt.Errorf("access error at %s: %w", path, err)
			return nil // Continue walking
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			// Skip excluded directories
			if s.isExcludedDirectory(path) {
				s.logger.Debugf("Skipping excluded directory: %s", path)
				return filepath.SkipDir
			}

			// Send directory to workers for processing
			jobs <- path
		}

		return nil
	})

	if err != nil && err != context.Canceled {
		errors <- fmt.Errorf("directory walk error: %w", err)
	}
}

// processDirectory scans a single directory for image files
func (s *Scanner) processDirectory(dir string, imagePaths *[]string, mu *sync.Mutex) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var dirImagePaths []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Directories are handled by the walker
		}

		filePath := filepath.Join(dir, entry.Name())

		// Check if file is a supported image format
		if s.isImageFile(filePath) {
			// Check file size if configured
			if s.config.MaxFileSize > 0 {
				info, err := entry.Info()
				if err != nil {
					s.logger.Debugf("Failed to get file info for %s: %v", filePath, err)
					continue
				}

				if info.Size() > s.config.MaxFileSize {
					s.logger.Debugf("Skipping large file: %s (%d bytes)", filePath, info.Size())
					continue
				}
			}

			dirImagePaths = append(dirImagePaths, filePath)
		}
	}

	// Add discovered images to the main list
	if len(dirImagePaths) > 0 {
		mu.Lock()
		*imagePaths = append(*imagePaths, dirImagePaths...)
		mu.Unlock()

		s.logger.Debugf("Found %d images in %s", len(dirImagePaths), dir)
	}

	return nil
}

// isImageFile checks if a file has a supported image extension
func (s *Scanner) isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, supportedExt := range s.config.SupportedFormats {
		if ext == supportedExt {
			return true
		}
	}
	return false
}

// isExcludedDirectory checks if a directory should be excluded from scanning
func (s *Scanner) isExcludedDirectory(path string) bool {
	dirName := filepath.Base(path)
	for _, excluded := range s.config.ExcludeDirs {
		if strings.EqualFold(dirName, excluded) {
			return true
		}
	}
	return false
}

// GetSupportedFormats returns the list of supported image formats
func (s *Scanner) GetSupportedFormats() []string {
	return s.config.SupportedFormats
}

// SetSupportedFormats updates the list of supported image formats
func (s *Scanner) SetSupportedFormats(formats []string) {
	s.config.SupportedFormats = formats
	s.logger.Infof("Updated supported formats: %v", formats)
}
