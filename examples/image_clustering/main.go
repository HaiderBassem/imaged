package main

import (
	"context"
	"fmt"
	"log"

	"github.com/HaiderBassem/imaged/internal/engine"
)

func main() {
	fmt.Println("ImageD Image Clustering Example")
	fmt.Println("===============================")

	// Create engine
	cfg := engine.DefaultConfig()
	cfg.IndexPath = "clustering.db"

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		log.Fatal("Failed to create engine:", err)
	}
	defer eng.Close()

	// Scan directory first
	ctx := context.Background()
	fmt.Println("Scanning directory...")
	err = eng.ScanFolder(ctx, "./photos", nil)
	if err != nil {
		log.Fatal("Scan failed:", err)
	}

	// For demonstration, we'll use duplicate detection as clustering
	fmt.Println("Finding similar images...")
	similarGroups, err := eng.FindNearDuplicates(0.7)
	if err != nil {
		log.Fatal("Failed to find similar images:", err)
	}

	// Display clustering results
	fmt.Printf("Found %d clusters/groups\n", len(similarGroups))

	for i, group := range similarGroups {
		fmt.Printf("\nCluster %d:\n", i+1)
		fmt.Printf("  Group ID: %s\n", group.GroupID)
		fmt.Printf("  Confidence: %.2f\n", group.Confidence)
		fmt.Printf("  Main Image: %s\n", group.MainImage)
		fmt.Printf("  Similar Images: %d\n", len(group.DuplicateIDs))

		if len(group.DuplicateIDs) > 0 {
			fmt.Printf("  Sample Files:\n")
			for j, dupID := range group.DuplicateIDs {
				if j < 3 {
					fmt.Printf("    - %s\n", dupID)
				}
			}
			if len(group.DuplicateIDs) > 3 {
				fmt.Printf("    ... and %d more\n", len(group.DuplicateIDs)-3)
			}
		}
	}

	fmt.Printf("\n Clustering completed! Found %d groups of similar images.\n", len(similarGroups))
}
