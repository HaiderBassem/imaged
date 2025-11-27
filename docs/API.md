# ImageD API Documentation

## Overview

ImageD is a professional image processing library for duplicate detection, quality analysis, and image management.

## Core Components

### Engine

The main entry point for image processing operations.

```go
// Create engine
cfg := engine.DefaultConfig()
eng, err := engine.NewEngine(cfg)

// Scan directory
err = eng.ScanFolder(ctx, "./photos", progressChan)

// Find duplicates
duplicates, err := eng.FindExactDuplicates()

// Analyze quality
quality, err := eng.RateImageQuality("image.jpg")

// Clean duplicates
report, err := eng.CleanDuplicates(options)


### Config

```go
type EngineConfig struct {
    IndexPath    string
    NumWorkers   int
    UseGPU       bool
    LogLevel     string
    MaxMemoryMB  int
}

// Default configuration
cfg := engine.DefaultConfig()

// High performance configuration  
cfg := engine.HighPerformanceConfig()

// Accuracy-focused configuration
cfg := engine.AccuracyConfig()
```

### Data Types

## ImageFingerprint

```go
type ImageFingerprint struct {
    ID          ImageID
    Metadata    ImageMetadata
    PHashes     PerceptualHashes
    Quality     ImageQuality
    CreatedAt   time.Time
}
```

## ImageQuality

```go
type ImageQuality struct {
    Sharpness   float64 // 0..1
    Noise       float64 // 0..1  
    Exposure    float64 // 0..1
    Contrast    float64 // 0..1
    FinalScore  float64 // 0..100
}
```


### Usage Examples

## Basic Scanning

```go
package main

import (
    "context"
    "github.com/HaiderBassem/imaged/internal/engine"
)

func main() {
    eng, _ := engine.NewEngine(engine.DefaultConfig())
    defer eng.Close()

    ctx := context.Background()
    eng.ScanFolder(ctx, "./photos", nil)
    
    duplicates, _ := eng.FindExactDuplicates()
    fmt.Printf("Found %d duplicates\n", len(duplicates))
}
```

## Quality Analysis

```go
quality, err := eng.RateImageQuality("photo.jpg")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Quality score: %.1f/100\n", quality.FinalScore)
fmt.Printf("Sharpness: %.2f\n", quality.Sharpness)
```

## Duplicate Cleaning

```go
options := api.CleanOptions{
    DryRun:          true,
    SelectionPolicy: api.PolicyHighestQuality,
    MoveDuplicates:  true,
    OutputDir:       "./duplicates",
}

report, err := eng.CleanDuplicates(options)
fmt.Printf("Moved %d files\n", report.MovedFiles)
```


## Error Handling

```go
eng, err := engine.NewEngine(cfg)
if err != nil {
    // Handle initialization error
}

err = eng.ScanFolder(ctx, path, nil)
if err != nil {
    // Handle scan error
}

// Check for specific errors
if errors.Is(err, api.ErrImageNotFound) {
    // Handle missing image
}
```