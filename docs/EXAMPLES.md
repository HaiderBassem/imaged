
# ImageD Examples

## Quick Start

### 1. Basic Scanning

```go
package main

import (
    "context"
    "fmt"
    "github.com/HaiderBassem/imaged/internal/engine"
)

func main() {
    eng, err := engine.NewEngine(engine.DefaultConfig())
    if err != nil {
        panic(err)
    }
    defer eng.Close()

    ctx := context.Background()
    err = eng.ScanFolder(ctx, "./photos", nil)
    if err != nil {
        panic(err)
    }

    duplicates, _ := eng.FindExactDuplicates()
    fmt.Printf("Found %d duplicate groups\n", len(duplicates))
}
```

## Quality Analysis

```go
quality, err := eng.RateImageQuality("image.jpg")
if err != nil {
    panic(err)
}

fmt.Printf("Quality Score: %.1f/100\n", quality.FinalScore)
if quality.FinalScore > 80 {
    fmt.Println("High quality image")
} else if quality.FinalScore > 60 {
    fmt.Println("Medium quality image") 
} else {
    fmt.Println("Low quality image")
}
```

## Custom Configuration

```go
cfg := engine.DefaultConfig()
cfg.IndexPath = "custom.db"
cfg.NumWorkers = 8
cfg.MaxMemoryMB = 2048

eng, err := engine.NewEngine(cfg)
```

## Progress Tracking

```go
progress := make(chan api.ScanProgress, 10)
go func() {
    for p := range progress {
        fmt.Printf("\rProgress: %.1f%%", p.Percentage)
    }
}()

err = eng.ScanFolder(ctx, "./photos", progress)
close(progress)
```


## Batch Processing

```go
// Process multiple directories
directories := []string{"./photos", "./archive", "./backup"}

for _, dir := range directories {
    fmt.Printf("Processing %s...\n", dir)
    err := eng.ScanFolder(ctx, dir, nil)
    if err != nil {
        fmt.Printf("Error processing %s: %v\n", dir, err)
    }
}
```


### Photo Library Cleanup


```go
// Comprehensive cleanup pipeline
func cleanupPhotoLibrary() error {
    eng, err := engine.NewEngine(engine.HighPerformanceConfig())
    if err != nil {
        return err
    }
    defer eng.Close()

    // Scan library
    ctx := context.Background()
    if err := eng.ScanFolder(ctx, "./photo-library", nil); err != nil {
        return err
    }

    // Find and remove duplicates
    options := api.CleanOptions{
        DryRun:          false,
        SelectionPolicy: api.PolicyHighestQuality,
        MoveDuplicates:  true,
        OutputDir:       "./duplicates",
    }

    report, err := eng.CleanDuplicates(options)
    if err != nil {
        return err
    }

    fmt.Printf("Cleaned %d duplicate files\n", report.MovedFiles)
    fmt.Printf("Freed ~%.1f MB\n", float64(report.FreedSpace)/1024/1024)

    return nil
}
```


### Quality-based Filtering

```go
// Filter images by quality threshold
func filterLowQualityImages(qualityThreshold float64) error {
    eng, err := engine.NewEngine(engine.DefaultConfig())
    if err != nil {
        return err
    }
    defer eng.Close()

    // This would require additional methods to get all fingerprints
    // and filter based on quality scores
    
    return nil
}
```

### CLI Integration

```bash
# Scan directory
imaged scan --path ./photos --index photos.db

# Find duplicates
imaged find-duplicates --index photos.db --threshold 0.9

# Clean duplicates
imaged clean --path ./photos --output ./duplicates --threshold 0.9
```
