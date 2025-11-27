package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/HaiderBassem/imaged/internal/engine"
	"github.com/urfave/cli/v2"
)

// ClusterCommand handles image clustering operations
func ClusterCommand(c *cli.Context) error {
	path := c.String("path")
	indexPath := c.String("index")
	threshold := c.Float64("threshold")

	if path == "" {
		return cli.Exit("Path is required", 1)
	}

	fmt.Printf("Clustering images in: %s\n", path)
	fmt.Printf("Similarity threshold: %.2f\n", threshold)

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

	// For now, we'll find duplicates as a clustering demonstration
	// In a full implementation, this would use proper clustering algorithms
	duplicates, err := eng.FindNearDuplicates(threshold)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to find similar images: %v", err), 1)
	}

	// Display clustering results
	fmt.Printf("\nClustering Results:\n")
	fmt.Printf("Found %d clusters/groups\n", len(duplicates))

	for i, group := range duplicates {
		fmt.Printf("\nCluster %d:\n", i+1)
		fmt.Printf("  Group ID: %s\n", group.GroupID)
		fmt.Printf("  Confidence: %.2f\n", group.Confidence)
		fmt.Printf("  Main Image: %s\n", group.MainImage)
		fmt.Printf("  Similar Images: %d\n", len(group.DuplicateIDs))

		if len(group.DuplicateIDs) > 0 {
			fmt.Printf("  Similar Files:\n")
			for j, dupID := range group.DuplicateIDs {
				if j < 5 { // Show only first 5 to avoid clutter
					fmt.Printf("    - %s\n", dupID)
				}
			}
			if len(group.DuplicateIDs) > 5 {
				fmt.Printf("    ... and %d more\n", len(group.DuplicateIDs)-5)
			}
		}
	}

	return nil
}
