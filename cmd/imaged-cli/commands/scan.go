package commands

import (
	"context"
	"fmt"

	"github.com/HaiderBassem/imaged/internal/engine"
	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/urfave/cli/v2"
)

// ScanCommand handles directory scanning operations
func ScanCommand(c *cli.Context) error {
	path := c.String("path")
	indexPath := c.String("index")
	workers := c.Int("workers")

	if path == "" {
		return cli.Exit("Path is required", 1)
	}

	fmt.Printf("Scanning directory: %s\n", path)
	fmt.Printf("Using index: %s\n", indexPath)
	fmt.Printf("Workers: %d\n", workers)

	// Create engine configuration
	cfg := engine.DefaultConfig()
	cfg.IndexPath = indexPath
	cfg.NumWorkers = workers

	// Initialize engine
	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create engine: %v", err), 1)
	}
	defer eng.Close()

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup progress channel
	progress := make(chan api.ScanProgress, 10)

	go displayScanProgress(progress)

	// Perform scan
	err = eng.ScanFolder(ctx, path, progress)
	close(progress)

	if err != nil {
		return cli.Exit(fmt.Sprintf("Scan failed: %v", err), 1)
	}

	// Get statistics
	stats, err := eng.GetStats()
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get stats: %v", err), 1)
	}

	fmt.Printf("\nScan completed successfully!\n")
	fmt.Printf("Total images: %d\n", stats.TotalImages)
	fmt.Printf("Total size: %.2f MB\n", float64(stats.TotalSizeBytes)/1024/1024)
	fmt.Printf("Average quality: %.1f/100\n", stats.AverageQuality)

	return nil
}

// displayScanProgress shows real-time scan progress
func displayScanProgress(progress <-chan api.ScanProgress) {
	for p := range progress {
		fmt.Printf("\rProgress: %.1f%% (%d/%d) - %s",
			p.Percentage, p.Current, p.Total, p.CurrentFile)
	}
	fmt.Println() // New line after progress completes
}
