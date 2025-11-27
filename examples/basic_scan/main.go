package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/HaiderBassem/imaged/pkg/engine"
)

func main() {
	// Create engine with default configuration
	cfg := engine.DefaultConfig()
	cfg.IndexPath = "example.db"

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		log.Fatal("Failed to create engine:", err)
	}
	defer eng.Close()

	// Scan a directory
	ctx := context.Background()
	fmt.Println("Starting scan...")

	start := time.Now()
	err = eng.ScanFolder(ctx, "/home/cpluspluser/Pictures/Images", nil)
	if err != nil {
		log.Fatal("Scan failed:", err)
	}

	fmt.Printf("Scan completed in %v\n", time.Since(start))

	// Find duplicates
	fmt.Println("Finding duplicates...")
	exactDuplicates, err := eng.FindExactDuplicates()
	if err != nil {
		log.Fatal("Failed to find duplicates:", err)
	}

	nearDuplicates, err := eng.FindNearDuplicates(0.8)
	if err != nil {
		log.Fatal("Failed to find near duplicates:", err)
	}

	fmt.Printf("Found %d exact duplicate groups\n", len(exactDuplicates))
	fmt.Printf("Found %d near-duplicate groups\n", len(nearDuplicates))

	// Show statistics
	stats, err := eng.GetStats()
	if err != nil {
		log.Fatal("Failed to get stats:", err)
	}

	fmt.Printf("Total images: %d\n", stats.TotalImages)
	fmt.Printf("Total size: %.2f MB\n", float64(stats.TotalSizeBytes)/1024/1024)
}
