package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/sirupsen/logrus"
)

// Generator creates various types of reports from scan results
type Generator struct {
	logger *logrus.Logger
}

// NewGenerator creates a new report generator
func NewGenerator() *Generator {
	return &Generator{
		logger: logrus.New(),
	}
}

// JSONReport generates a JSON format report
func (g *Generator) JSONReport(scanReport *api.ScanReport, outputPath string) error {
	data, err := json.MarshalIndent(scanReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	g.logger.Infof("JSON report saved to: %s", outputPath)
	return nil
}

// TextReport generates a human-readable text report
func (g *Generator) TextReport(scanReport *api.ScanReport, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create text report: %w", err)
	}
	defer file.Close()

	// Generate comprehensive text report
	report := g.generateTextContent(scanReport)
	if _, err := file.WriteString(report); err != nil {
		return fmt.Errorf("failed to write text report: %w", err)
	}

	g.logger.Infof("Text report saved to: %s", outputPath)
	return nil
}

// generateTextContent creates the content for text reports
func (g *Generator) generateTextContent(report *api.ScanReport) string {
	var content string

	content += "IMAGE DEDUPLICATION REPORT\n"
	content += "==========================\n\n"

	// Summary section
	content += "SUMMARY\n"
	content += "-------\n"
	content += fmt.Sprintf("Scan ID: %s\n", report.ScanID)
	content += fmt.Sprintf("Scan Date: %s\n", report.StartedAt.Format("2006-01-02 15:04:05"))
	content += fmt.Sprintf("Duration: %v\n", report.ScanDuration.Round(time.Second))
	content += fmt.Sprintf("Total Files: %d\n", report.TotalFiles)
	content += fmt.Sprintf("Processed Images: %d\n", report.ProcessedImages)
	content += fmt.Sprintf("Skipped Files: %d\n", report.SkippedFiles)
	content += fmt.Sprintf("Exact Duplicates: %d groups\n", report.ExactDuplicateCount)
	content += fmt.Sprintf("Near Duplicates: %d groups\n", report.NearDuplicateCount)
	content += fmt.Sprintf("Total Clusters: %d\n\n", len(report.Clusters))

	// Duplicate groups section
	if len(report.Groups) > 0 {
		content += "DUPLICATE GROUPS\n"
		content += "----------------\n"

		for i, group := range report.Groups {
			content += fmt.Sprintf("Group %d: %s\n", i+1, group.GroupID)
			content += fmt.Sprintf("  Reason: %s\n", group.Reason)
			content += fmt.Sprintf("  Confidence: %.2f\n", group.Confidence)
			content += fmt.Sprintf("  Main Image: %s\n", group.MainImage)
			content += fmt.Sprintf("  Duplicates: %d files\n", len(group.DuplicateIDs))

			if len(group.DuplicateIDs) > 0 {
				content += "  Duplicate Files:\n"
				for _, dupID := range group.DuplicateIDs {
					content += fmt.Sprintf("    - %s\n", dupID)
				}
			}
			content += "\n"
		}
	}

	// Clusters section
	if len(report.Clusters) > 0 {
		content += "IMAGE CLUSTERS\n"
		content += "--------------\n"

		for i, cluster := range report.Clusters {
			content += fmt.Sprintf("Cluster %d: %s\n", i+1, cluster.ClusterID)
			if cluster.Name != "" {
				content += fmt.Sprintf("  Name: %s\n", cluster.Name)
			}
			content += fmt.Sprintf("  Images: %d files\n", len(cluster.Images))
			content += "\n"
		}
	}

	// Recommendations section
	content += "RECOMMENDATIONS\n"
	content += "---------------\n"
	if report.ExactDuplicateCount > 0 {
		content += fmt.Sprintf("- Remove %d exact duplicate groups to save space\n", report.ExactDuplicateCount)
	}
	if report.NearDuplicateCount > 0 {
		content += fmt.Sprintf("- Review %d near-duplicate groups for similar images\n", report.NearDuplicateCount)
	}
	if len(report.Clusters) > 0 {
		content += fmt.Sprintf("- Organize images into %d thematic clusters\n", len(report.Clusters))
	}

	content += fmt.Sprintf("\nReport generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return content
}
