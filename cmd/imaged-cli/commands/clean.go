package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/HaiderBassem/imaged/pkg/engine"
	"github.com/urfave/cli/v2"
)

// CleanCommand handles duplicate cleaning operations
func CleanCommand(c *cli.Context) error {
	path := c.String("path")
	indexPath := c.String("index")
	outputDir := c.String("output")
	threshold := c.Float64("threshold")
	dryRun := c.Bool("dry-run")
	move := c.Bool("move")

	if path == "" {
		return cli.Exit("Path is required", 1)
	}

	fmt.Printf("Cleaning directory: %s\n", path)
	if dryRun {
		fmt.Println("DRY RUN MODE - No files will be modified")
	}

	cfg := engine.DefaultConfig()
	cfg.IndexPath = indexPath

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create engine: %v", err), 1)
	}
	defer eng.Close()

	// Check if index exists, scan if not
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		fmt.Println("Index not found, scanning directory first...")
		ctx := context.Background()
		if err := eng.ScanFolder(ctx, path, nil); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to scan directory: %v", err), 1)
		}
	}

	// Setup clean options
	options := api.CleanOptions{
		DryRun:                 dryRun,
		SelectionPolicy:        api.PolicyHighestQuality,
		MinQualityScore:        50.0,
		MaxSimilarityThreshold: threshold,
		MoveDuplicates:         move,
		OutputDir:              outputDir,
	}

	// Perform cleaning
	report, err := eng.CleanDuplicates(options)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Clean failed: %v", err), 1)
	}

	// Display results
	fmt.Printf("\nClean operation completed:\n")
	fmt.Printf("  Total groups processed: %d\n", report.TotalProcessed)
	fmt.Printf("  Files moved/deleted: %d\n", report.MovedFiles)
	fmt.Printf("  Storage freed: %s\n", formatBytes(report.FreedSpace))
	fmt.Printf("  Errors: %d\n", report.Errors)

	if dryRun {
		fmt.Println("\nThis was a dry run. Run without --dry-run to actually clean files.")
	}

	return nil
}

// formatBytes converts bytes to human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
