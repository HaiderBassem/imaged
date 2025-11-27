package report

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/sirupsen/logrus"
)

// TextReportGenerator generates human-readable text reports
type TextReportGenerator struct {
	logger *logrus.Logger
}

// NewTextReportGenerator creates a new text report generator
func NewTextReportGenerator() *TextReportGenerator {
	return &TextReportGenerator{
		logger: logrus.New(),
	}
}

// Generate generates a comprehensive text report
func (t *TextReportGenerator) Generate(scanReport *api.ScanReport, outputPath string) error {
	content := t.generateReportContent(scanReport)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create text report: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write text report: %w", err)
	}

	t.logger.Infof("Text report saved to: %s", outputPath)
	return nil
}

// generateReportContent creates the text content for the report
func (t *TextReportGenerator) generateReportContent(report *api.ScanReport) string {
	var sb strings.Builder

	// Header
	sb.WriteString(t.generateHeader(report))
	sb.WriteString("\n\n")

	// Summary
	sb.WriteString(t.generateSummary(report))
	sb.WriteString("\n\n")

	// Duplicate Analysis
	if len(report.Groups) > 0 {
		sb.WriteString(t.generateDuplicateAnalysis(report))
		sb.WriteString("\n\n")
	}

	// Clustering Analysis
	if len(report.Clusters) > 0 {
		sb.WriteString(t.generateClusteringAnalysis(report))
		sb.WriteString("\n\n")
	}

	// Recommendations
	sb.WriteString(t.generateRecommendations(report))
	sb.WriteString("\n\n")

	// Footer
	sb.WriteString(t.generateFooter())

	return sb.String()
}

// generateHeader creates the report header
func (t *TextReportGenerator) generateHeader(report *api.ScanReport) string {
	return fmt.Sprintf(`IMAGE DEDUPLICATION REPORT
==========================
Scan ID: %s
Generated: %s
Scan Date: %s - %s
Duration: %v`,
		report.ScanID,
		time.Now().Format("2006-01-02 15:04:05"),
		report.StartedAt.Format("2006-01-02 15:04:05"),
		report.CompletedAt.Format("2006-01-02 15:04:05"),
		report.ScanDuration.Round(time.Second),
	)
}

// generateSummary creates the summary section
func (t *TextReportGenerator) generateSummary(report *api.ScanReport) string {
	return fmt.Sprintf(`SUMMARY
-------
Total Files Scanned: %d
Images Processed: %d
Files Skipped: %d
Exact Duplicate Groups: %d
Near-Duplicate Groups: %d
Image Clusters: %d`,
		report.TotalFiles,
		report.ProcessedImages,
		report.SkippedFiles,
		report.ExactDuplicateCount,
		report.NearDuplicateCount,
		len(report.Clusters),
	)
}

// generateDuplicateAnalysis creates the duplicate analysis section
func (t *TextReportGenerator) generateDuplicateAnalysis(report *api.ScanReport) string {
	var sb strings.Builder

	sb.WriteString("DUPLICATE ANALYSIS\n")
	sb.WriteString("------------------\n")

	exactGroups := 0
	nearGroups := 0

	for i, group := range report.Groups {
		sb.WriteString(fmt.Sprintf("Group %d:\n", i+1))
		sb.WriteString(fmt.Sprintf("  Type: %s\n", strings.ToUpper(group.Reason)))
		sb.WriteString(fmt.Sprintf("  Confidence: %.2f\n", group.Confidence))
		sb.WriteString(fmt.Sprintf("  Main Image: %s\n", group.MainImage))
		sb.WriteString(fmt.Sprintf("  Duplicates: %d files\n", len(group.DuplicateIDs)))

		if group.Reason == "exact" {
			exactGroups++
		} else {
			nearGroups++
		}

		// Show first few duplicates
		if len(group.DuplicateIDs) > 0 {
			sb.WriteString("  Duplicate Files:\n")
			for j, dupID := range group.DuplicateIDs {
				if j < 3 { // Show only first 3 to avoid clutter
					sb.WriteString(fmt.Sprintf("    - %s\n", dupID))
				}
			}
			if len(group.DuplicateIDs) > 3 {
				sb.WriteString(fmt.Sprintf("    ... and %d more\n", len(group.DuplicateIDs)-3))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Summary: %d exact groups, %d near-duplicate groups\n", exactGroups, nearGroups))

	return sb.String()
}

// generateClusteringAnalysis creates the clustering analysis section
func (t *TextReportGenerator) generateClusteringAnalysis(report *api.ScanReport) string {
	var sb strings.Builder

	sb.WriteString("IMAGE CLUSTERS\n")
	sb.WriteString("--------------\n")

	for i, cluster := range report.Clusters {
		sb.WriteString(fmt.Sprintf("Cluster %d: %s\n", i+1, cluster.ClusterID))
		if cluster.Name != "" {
			sb.WriteString(fmt.Sprintf("  Name: %s\n", cluster.Name))
		}
		sb.WriteString(fmt.Sprintf("  Images: %d files\n", len(cluster.Images)))

		// Show first few images in cluster
		if len(cluster.Images) > 0 {
			sb.WriteString("  Sample Images:\n")
			for j, imgID := range cluster.Images {
				if j < 3 { // Show only first 3
					sb.WriteString(fmt.Sprintf("    - %s\n", imgID))
				}
			}
			if len(cluster.Images) > 3 {
				sb.WriteString(fmt.Sprintf("    ... and %d more\n", len(cluster.Images)-3))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// generateRecommendations creates actionable recommendations
func (t *TextReportGenerator) generateRecommendations(report *api.ScanReport) string {
	var sb strings.Builder

	sb.WriteString("RECOMMENDATIONS\n")
	sb.WriteString("---------------\n")

	if report.ExactDuplicateCount > 0 {
		sb.WriteString(fmt.Sprintf("✅ Remove %d exact duplicate groups to save storage space\n", report.ExactDuplicateCount))
		sb.WriteString("   Command: imaged clean --path <directory> --threshold 1.0\n\n")
	}

	if report.NearDuplicateCount > 0 {
		sb.WriteString(fmt.Sprintf("🔍 Review %d near-duplicate groups for similar images\n", report.NearDuplicateCount))
		sb.WriteString("   Command: imaged find-duplicates --threshold 0.8\n\n")
	}

	if len(report.Clusters) > 0 {
		sb.WriteString(fmt.Sprintf("📁 Organize images into %d thematic clusters\n", len(report.Clusters)))
		sb.WriteString("   Use clusters to create organized folder structure\n\n")
	}

	if report.SkippedFiles > 0 {
		sb.WriteString(fmt.Sprintf("⚠️  %d files were skipped during processing\n", report.SkippedFiles))
		sb.WriteString("   Check file formats and permissions\n\n")
	}

	sb.WriteString("NEXT STEPS:\n")
	sb.WriteString("1. Run 'imaged clean' to remove duplicates\n")
	sb.WriteString("2. Use 'imaged quality' to analyze image quality\n")
	sb.WriteString("3. Consider organizing images by clusters\n")

	return sb.String()
}

// generateFooter creates the report footer
func (t *TextReportGenerator) generateFooter() string {
	return fmt.Sprintf(`---
Report generated by ImageD
Generated at: %s
For more information: https://github.com/yourusername/imaged`,
		time.Now().Format("2006-01-02 15:04:05"),
	)
}

// GenerateBrief generates a brief summary report
func (t *TextReportGenerator) GenerateBrief(scanReport *api.ScanReport, outputPath string) error {
	content := fmt.Sprintf(`QUICK SCAN REPORT
=================

Scan Results:
- Total Files: %d
- Images Processed: %d
- Exact Duplicates: %d groups
- Near Duplicates: %d groups
- Scan Duration: %v

Generated: %s`,
		scanReport.TotalFiles,
		scanReport.ProcessedImages,
		scanReport.ExactDuplicateCount,
		scanReport.NearDuplicateCount,
		scanReport.ScanDuration.Round(time.Second),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	return os.WriteFile(outputPath, []byte(content), 0644)
}
