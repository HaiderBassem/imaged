package engine

import "github.com/HaiderBassem/imaged/internal/quality"

// DefaultConfig returns sensible default configuration for the engine
func DefaultConfig() EngineConfig {
	return EngineConfig{
		IndexPath:   "imaged.db",
		NumWorkers:  4,
		UseGPU:      false,
		LogLevel:    "info",
		MaxMemoryMB: 1024,
		HashConfig: HashConfig{
			ComputeAHash: true,
			ComputePHash: true,
			ComputeDHash: true,
			ComputeWHash: false,
			HashSize:     8,
		},
		QualityConfig: quality.DefaultConfig(),
	}
}

// HighPerformanceConfig returns configuration optimized for performance
func HighPerformanceConfig() EngineConfig {
	cfg := DefaultConfig()
	cfg.NumWorkers = 8
	cfg.MaxMemoryMB = 2048
	cfg.HashConfig.ComputeWHash = true
	return cfg
}

// AccuracyConfig returns configuration optimized for accuracy over performance
func AccuracyConfig() EngineConfig {
	cfg := DefaultConfig()
	cfg.HashConfig.HashSize = 16
	cfg.QualityConfig.DetailedAnalysis = true
	cfg.NumWorkers = 2 // Fewer workers for more accurate analysis
	return cfg
}

// FastScanConfig returns configuration optimized for quick scanning
func FastScanConfig() EngineConfig {
	cfg := DefaultConfig()
	cfg.NumWorkers = 1
	cfg.HashConfig.ComputePHash = false
	cfg.HashConfig.ComputeWHash = false
	cfg.QualityConfig.DetailedAnalysis = false
	return cfg
}
