package commands

import (
	"fmt"

	"github.com/HaiderBassem/imaged/internal/engine"
	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/urfave/cli/v2"
)

// FindDuplicatesCommand handles duplicate detection operations
func FindDuplicatesCommand(c *cli.Context) error {
	indexPath := c.String("index")
	threshold := c.Float64("threshold")
	exactOnly := c.Bool("exact-only")

	fmt.Printf("Finding duplicates in index: %s\n", indexPath)
	if exactOnly {
		fmt.Println("Mode: Exact duplicates only")
	} else {
		fmt.Printf("Similarity threshold: %.2f\n", threshold)
	}

	cfg := engine.DefaultConfig()
	cfg.IndexPath = indexPath

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create engine: %v", err), 1)
	}
	defer eng.Close()

	var exactGroups, nearGroups []api.DuplicateGroup

	// Find exact duplicates
	exactGroups, err = eng.FindExactDuplicates()
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to find exact duplicates: %v", err), 1)
	}

	if !exactOnly {
		// Find near duplicates
		nearGroups, err = eng.FindNearDuplicates(threshold)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to find near duplicates: %v", err), 1)
		}
	}

	// Display results
	displayDuplicateResults(exactGroups, nearGroups, exactOnly)

	return nil
}

// displayDuplicateResults shows duplicate detection results
func displayDuplicateResults(
	exactGroups, nearGroups []api.DuplicateGroup,
	exactOnly bool,
) {
	fmt.Printf("\nDUPLICATE DETECTION RESULTS\n\n")

	// Exact duplicates
	if len(exactGroups) > 0 {
		fmt.Printf("EXACT DUPLICATES (%d groups):\n\n", len(exactGroups))

		totalExactFiles := 0
		for i, group := range exactGroups {
			fmt.Printf("Group %d:\n", i+1)
			fmt.Printf("  Main Image: %s\n", group.MainImage)
			fmt.Printf("  Duplicates: %d files\n", len(group.DuplicateIDs))

			for j, dupID := range group.DuplicateIDs {
				if j < 3 {
					fmt.Printf("    - %s\n", dupID)
				}
			}
			if len(group.DuplicateIDs) > 3 {
				fmt.Printf("    ... and %d more\n", len(group.DuplicateIDs)-3)
			}
			fmt.Println()

			totalExactFiles += len(group.DuplicateIDs)
		}

		estimatedSavings := totalExactFiles * 5
		fmt.Printf("Total exact duplicate files: %d\n", totalExactFiles)
		fmt.Printf("Estimated storage savings: ~%d MB\n", estimatedSavings)
		fmt.Println()
	} else {
		fmt.Printf("No exact duplicates found.\n\n")
	}

	// Near duplicates
	if len(nearGroups) > 0 {
		fmt.Printf("NEAR DUPLICATES (%d groups):\n\n", len(nearGroups))

		totalNearFiles := 0
		for i, group := range nearGroups {
			fmt.Printf("Group %d (confidence: %.2f):\n", i+1, group.Confidence)
			fmt.Printf("  Main Image: %s\n", group.MainImage)
			fmt.Printf("  Similar Images: %d files\n", len(group.DuplicateIDs))

			for j, dupID := range group.DuplicateIDs {
				if j < 2 {
					fmt.Printf("    - %s\n", dupID)
				}
			}
			if len(group.DuplicateIDs) > 2 {
				fmt.Printf("    ... and %d more\n", len(group.DuplicateIDs)-2)
			}
			fmt.Println()

			totalNearFiles += len(group.DuplicateIDs)
		}

		fmt.Printf("Total near-duplicate files: %d\n", totalNearFiles)
		fmt.Println()
	} else if !exactOnly {
		fmt.Printf("No near duplicates found.\n\n")
	}

	// Summary
	totalGroups := len(exactGroups) + len(nearGroups)
	if totalGroups > 0 {
		fmt.Printf("SUMMARY:\n")
		fmt.Printf("--------\n")
		fmt.Printf("Total duplicate groups: %d\n", totalGroups)
		fmt.Printf("Exact duplicate groups: %d\n", len(exactGroups))
		fmt.Printf("Near-duplicate groups: %d\n", len(nearGroups))

		if len(exactGroups) > 0 {
			fmt.Printf("\nRun 'imaged clean' to remove exact duplicates.\n")
		}
		if len(nearGroups) > 0 {
			fmt.Printf("Review near-duplicates manually.\n")
		}
	} else {
		fmt.Printf("No duplicates found.\n")
	}
}
