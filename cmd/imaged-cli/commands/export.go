package commands

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/HaiderBassem/imaged/internal/engine"
	"github.com/HaiderBassem/imaged/internal/report"
	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/urfave/cli/v2"
)

// ExportCommand handles report export operations
func ExportCommand(c *cli.Context) error {
	indexPath := c.String("index")
	format := c.String("format")
	output := c.String("output")

	fmt.Printf("Exporting report from index: %s\n", indexPath)
	fmt.Printf("Format: %s\n", format)
	fmt.Printf("Output: %s\n", output)

	cfg := engine.DefaultConfig()
	cfg.IndexPath = indexPath

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create engine: %v", err), 1)
	}
	defer eng.Close()

	// Get scan statistics (in a real implementation, you'd get actual scan data)
	stats, err := eng.GetStats()
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get statistics: %v", err), 1)
	}

	// Create a mock scan report for demonstration
	scanReport := &api.ScanReport{
		ScanID:              "export_" + time.Now().Format("20060102_150405"),
		TotalFiles:          int(stats.TotalImages),
		ProcessedImages:     int(stats.TotalImages),
		SkippedFiles:        0,
		ExactDuplicateCount: 0,         // Would be calculated from actual data
		NearDuplicateCount:  0,         // Would be calculated from actual data
		ScanDuration:        time.Hour, // Example
		StartedAt:           time.Now().Add(-time.Hour),
		CompletedAt:         time.Now(),
	}

	// Generate report based on format
	switch format {
	case "json":
		outputPath := addExtension(output, "json")
		generator := report.NewJSONReportGenerator()
		if err := generator.Generate(scanReport, outputPath); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to generate JSON report: %v", err), 1)
		}

	case "html":
		outputPath := addExtension(output, "html")
		generator := report.NewHTMLReportGenerator()
		if err := generator.Generate(scanReport, outputPath); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to generate HTML report: %v", err), 1)
		}

	case "text":
		outputPath := addExtension(output, "txt")
		generator := report.NewTextReportGenerator()
		if err := generator.Generate(scanReport, outputPath); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to generate text report: %v", err), 1)
		}

	default:
		return cli.Exit(fmt.Sprintf("Unsupported format: %s", format), 1)
	}

	fmt.Printf("✅ Report exported successfully to: %s\n", output)
	return nil
}

// addExtension adds file extension if not present
func addExtension(path, ext string) string {
	if filepath.Ext(path) == "" {
		return path + "." + ext
	}
	return path
}
