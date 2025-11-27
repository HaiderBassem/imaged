package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/HaiderBassem/imaged/pkg/engine"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "imaged",
		Version: "1.0.0",
		Usage:   "Professional image deduplication and management tool",
		Commands: []*cli.Command{
			{
				Name:  "scan",
				Usage: "Scan a directory and index images",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Usage:    "Directory path to scan",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "index",
						Aliases: []string{"i"},
						Usage:   "Index database path",
						Value:   "imaged.db",
					},
					&cli.IntFlag{
						Name:    "workers",
						Aliases: []string{"w"},
						Usage:   "Number of worker threads",
						Value:   4,
					},
				},
				Action: scanCommand,
			},
			{
				Name:  "find-duplicates",
				Usage: "Find duplicate images",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "index",
						Aliases: []string{"i"},
						Usage:   "Index database path",
						Value:   "imaged.db",
					},
					&cli.Float64Flag{
						Name:    "threshold",
						Aliases: []string{"t"},
						Usage:   "Similarity threshold (0.0-1.0)",
						Value:   0.9,
					},
				},
				Action: findDuplicatesCommand,
			},
			{
				Name:  "clean",
				Usage: "Clean duplicate images",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Usage:    "Directory path to clean",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "index",
						Aliases: []string{"i"},
						Usage:   "Index database path",
						Value:   "imaged.db",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output directory for duplicates",
						Value:   "duplicates",
					},
					&cli.Float64Flag{
						Name:    "threshold",
						Aliases: []string{"t"},
						Usage:   "Similarity threshold",
						Value:   0.9,
					},
					&cli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"d"},
						Usage:   "Show what would be done without actually doing it",
					},
					&cli.BoolFlag{
						Name:  "move",
						Usage: "Move duplicates instead of deleting them",
						Value: true,
					},
				},
				Action: cleanCommand,
			},
			{
				Name:  "quality",
				Usage: "Analyze image quality",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "image",
						Aliases:  []string{"i"},
						Usage:    "Image file to analyze",
						Required: true,
					},
				},
				Action: qualityCommand,
			},
			{
				Name:  "stats",
				Usage: "Show database statistics",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "index",
						Aliases: []string{"i"},
						Usage:   "Index database path",
						Value:   "imaged.db",
					},
				},
				Action: statsCommand,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func scanCommand(c *cli.Context) error {
	path := c.String("path")
	indexPath := c.String("index")
	workers := c.Int("workers")

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
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	setupInterruptHandler(cancel)

	// Setup progress reporting
	progress := make(chan api.ScanProgress, 10)
	go displayProgress(progress)

	// Perform scan
	err = eng.ScanFolder(ctx, path, progress)
	close(progress)

	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	fmt.Println("\nScan completed successfully!")
	return nil
}

func findDuplicatesCommand(c *cli.Context) error {
	indexPath := c.String("index")
	threshold := c.Float64("threshold")

	fmt.Printf("Finding duplicates in index: %s\n", indexPath)
	fmt.Printf("Similarity threshold: %.2f\n", threshold)

	cfg := engine.DefaultConfig()
	cfg.IndexPath = indexPath

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	// Find exact duplicates
	exactGroups, err := eng.FindExactDuplicates()
	if err != nil {
		return fmt.Errorf("failed to find exact duplicates: %w", err)
	}

	// Find near duplicates
	nearGroups, err := eng.FindNearDuplicates(threshold)
	if err != nil {
		return fmt.Errorf("failed to find near duplicates: %w", err)
	}

	// Display results
	fmt.Printf("\nExact duplicates: %d groups\n", len(exactGroups))
	for _, group := range exactGroups {
		fmt.Printf("  Group %s: %d duplicates\n", group.GroupID, len(group.DuplicateIDs))
	}

	fmt.Printf("\nNear duplicates: %d groups\n", len(nearGroups))
	for _, group := range nearGroups {
		fmt.Printf("  Group %s: %d duplicates (confidence: %.2f)\n",
			group.GroupID, len(group.DuplicateIDs), group.Confidence)
	}

	totalDuplicates := len(exactGroups) + len(nearGroups)
	fmt.Printf("\nTotal duplicate groups found: %d\n", totalDuplicates)

	return nil
}

func cleanCommand(c *cli.Context) error {
	path := c.String("path")
	indexPath := c.String("index")
	outputDir := c.String("output")
	threshold := c.Float64("threshold")
	dryRun := c.Bool("dry-run")
	move := c.Bool("move")

	fmt.Printf("Cleaning directory: %s\n", path)
	if dryRun {
		fmt.Println("DRY RUN MODE - No files will be modified")
	}

	cfg := engine.DefaultConfig()
	cfg.IndexPath = indexPath

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	// First scan the directory if needed
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		fmt.Println("Index not found, scanning directory first...")
		ctx := context.Background()
		if err := eng.ScanFolder(ctx, path, nil); err != nil {
			return fmt.Errorf("failed to scan directory: %w", err)
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
		return fmt.Errorf("clean failed: %w", err)
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

func qualityCommand(c *cli.Context) error {
	imagePath := c.String("image")

	fmt.Printf("Analyzing image quality: %s\n", imagePath)

	cfg := engine.DefaultConfig()
	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	quality, err := eng.RateImageQuality(imagePath)
	if err != nil {
		return fmt.Errorf("failed to analyze quality: %w", err)
	}

	// Display quality metrics
	fmt.Printf("\nQuality Analysis Results:\n")
	fmt.Printf("  Overall Score: %.1f/100\n", quality.FinalScore)
	fmt.Printf("  Sharpness: %.2f\n", quality.Sharpness)
	fmt.Printf("  Noise: %.2f\n", quality.Noise)
	fmt.Printf("  Exposure: %.2f\n", quality.Exposure)
	fmt.Printf("  Contrast: %.2f\n", quality.Contrast)
	fmt.Printf("  Compression: %.2f\n", quality.Compression)
	fmt.Printf("  Color Cast: %.2f\n", quality.ColorCast)

	// Provide recommendations
	fmt.Printf("\nRecommendations:\n")
	if quality.Sharpness < 0.3 {
		fmt.Printf("    Image is blurry (sharpness: %.2f)\n", quality.Sharpness)
	}
	if quality.Noise > 0.7 {
		fmt.Printf("    High noise level (%.2f)\n", quality.Noise)
	}
	if math.Abs(quality.Exposure-0.5) > 0.3 {
		fmt.Printf("    Exposure issues (%.2f)\n", quality.Exposure)
	}
	if quality.FinalScore >= 80 {
		fmt.Printf("   Excellent quality image\n")
	} else if quality.FinalScore >= 60 {
		fmt.Printf("    Average quality image\n")
	} else {
		fmt.Printf("   Poor quality image\n")
	}

	return nil
}

func statsCommand(c *cli.Context) error {
	indexPath := c.String("index")

	fmt.Printf("Database statistics: %s\n", indexPath)

	cfg := engine.DefaultConfig()
	cfg.IndexPath = indexPath

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	stats, err := eng.GetStats()
	if err != nil {
		return fmt.Errorf("failed to get statistics: %w", err)
	}

	fmt.Printf("\nIndex Statistics:\n")
	fmt.Printf("  Total images: %d\n", stats.TotalImages)
	fmt.Printf("  Total size: %s\n", formatBytes(stats.TotalSizeBytes))
	fmt.Printf("  Index size: %s\n", formatBytes(stats.IndexSizeBytes))
	fmt.Printf("  Average quality: %.1f/100\n", stats.AverageQuality)
	fmt.Printf("  Duplicate groups: %d\n", stats.DuplicateGroups)

	return nil
}

func displayProgress(progress <-chan api.ScanProgress) {
	var lastPercentage int
	for p := range progress {
		currentPercentage := int(p.Percentage)
		if currentPercentage != lastPercentage {
			fmt.Printf("\rProgress: %d%% (%d/%d) - %s",
				currentPercentage, p.Current, p.Total, filepath.Base(p.CurrentFile))
			lastPercentage = currentPercentage
		}
	}
	fmt.Println() // New line after progress completes
}

func setupInterruptHandler(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, stopping...")
		cancel()
	}()
}

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
