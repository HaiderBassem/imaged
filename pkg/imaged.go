package imaged

// Main package exports for easy library usage

import (
	"context"
	"time"

	"github.com/HaiderBassem/imaged/internal/engine"
	"github.com/HaiderBassem/imaged/internal/quality"
	"github.com/HaiderBassem/imaged/internal/scanner"
	"github.com/HaiderBassem/imaged/internal/similarity"
	"github.com/HaiderBassem/imaged/pkg/api"
)

// Re-export commonly used types and functions for public API

// Engine creation with different configurations
var (
	NewEngine             = engine.NewEngine
	DefaultConfig         = engine.DefaultConfig
	HighPerformanceConfig = engine.HighPerformanceConfig
	AccuracyConfig        = engine.AccuracyConfig
)

// Common types
type (
	EngineConfig     = engine.EngineConfig
	ImageFingerprint = api.ImageFingerprint
	ImageQuality     = api.ImageQuality
	DuplicateGroup   = api.DuplicateGroup
	ScanReport       = api.ScanReport
	CleanOptions     = api.CleanOptions
	SelectionPolicy  = api.SelectionPolicy
)

// Constants
const (
	PolicyHighestQuality    = api.PolicyHighestQuality
	PolicyHighestResolution = api.PolicyHighestResolution
	PolicyBestExposure      = api.PolicyBestExposure
	PolicyOldest            = api.PolicyOldest
	PolicyNewest            = api.PolicyNewest
)

// Scanner functionality
var (
	NewScanner           = scanner.NewScanner
	DefaultScannerConfig = scanner.DefaultConfig
)

// Quality analysis
var (
	NewQualityAnalyzer   = quality.NewAnalyzer
	DefaultQualityConfig = quality.DefaultConfig
)

// Similarity comparison
var (
	NewComparator = similarity.NewComparator
)

// Utility functions for common operations

// QuickScan performs a quick scan and returns basic statistics
func QuickScan(directoryPath string) (*api.ScanReport, error) {
	cfg := engine.DefaultConfig()
	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return nil, err
	}
	defer eng.Close()

	// Perform scan
	if err := eng.ScanFolder(context.Background(), directoryPath, nil); err != nil {
		return nil, err
	}

	// Get basic report
	stats, err := eng.GetStats()
	if err != nil {
		return nil, err
	}

	return &api.ScanReport{
		TotalFiles:      int(stats.TotalImages),
		ProcessedImages: int(stats.TotalImages),
		ScanDuration:    time.Since(time.Now()),
	}, nil
}

// FindDuplicatesQuick quickly finds duplicates in a directory
func FindDuplicatesQuick(directoryPath string, similarityThreshold float64) ([]api.DuplicateGroup, error) {
	cfg := engine.FastScanConfig()
	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return nil, err
	}
	defer eng.Close()

	// Scan directory
	if err := eng.ScanFolder(context.Background(), directoryPath, nil); err != nil {
		return nil, err
	}

	// Find duplicates
	exactDuplicates, err := eng.FindExactDuplicates()
	if err != nil {
		return nil, err
	}

	nearDuplicates, err := eng.FindNearDuplicates(similarityThreshold)
	if err != nil {
		return nil, err
	}

	return append(exactDuplicates, nearDuplicates...), nil
}

// AnalyzeImageQuality provides quick quality analysis for a single image
func AnalyzeImageQuality(imagePath string) (*api.ImageQuality, error) {
	eng, err := engine.NewEngine(engine.DefaultConfig())
	if err != nil {
		return nil, err
	}
	defer eng.Close()

	return eng.RateImageQuality(imagePath)
}
