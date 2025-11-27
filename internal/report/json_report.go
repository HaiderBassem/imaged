package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/sirupsen/logrus"
)

// JSONReportGenerator generates JSON format reports
type JSONReportGenerator struct {
	logger *logrus.Logger
}

// NewJSONReportGenerator creates a new JSON report generator
func NewJSONReportGenerator() *JSONReportGenerator {
	return &JSONReportGenerator{
		logger: logrus.New(),
	}
}

// Generate generates a comprehensive JSON report
func (j *JSONReportGenerator) Generate(scanReport *api.ScanReport, outputPath string) error {
	// Create enhanced report structure
	enhancedReport := j.enhanceReport(scanReport)

	data, err := json.MarshalIndent(enhancedReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	j.logger.Infof("JSON report saved to: %s", outputPath)
	return nil
}

// enhanceReport adds additional information to the basic scan report
func (j *JSONReportGenerator) enhanceReport(scanReport *api.ScanReport) *EnhancedReport {
	enhanced := &EnhancedReport{
		ScanReport:  scanReport,
		GeneratedAt: time.Now(),
		Version:     "1.0",
	}

	// Calculate additional statistics
	enhanced.Statistics = j.calculateStatistics(scanReport)

	// Add recommendations
	enhanced.Recommendations = j.generateRecommendations(scanReport)

	return enhanced
}

// EnhancedReport represents an enhanced report with additional information
type EnhancedReport struct {
	*api.ScanReport
	GeneratedAt     time.Time         `json:"generated_at"`
	Version         string            `json:"version"`
	Statistics      *ReportStatistics `json:"statistics"`
	Recommendations []*Recommendation `json:"recommendations"`
}

// ReportStatistics contains detailed statistics about the scan
type ReportStatistics struct {
	TotalSizeMB         float64        `json:"total_size_mb"`
	AverageFileSizeMB   float64        `json:"average_file_size_mb"`
	AverageQuality      float64        `json:"average_quality"`
	QualityDistribution map[string]int `json:"quality_distribution"`
	FormatDistribution  map[string]int `json:"format_distribution"`
	SizeDistribution    map[string]int `json:"size_distribution"`
}

// Recommendation represents an action recommendation
type Recommendation struct {
	Type        string  `json:"type"`
	Priority    string  `json:"priority"` // low, medium, high, critical
	Description string  `json:"description"`
	Action      string  `json:"action"`
	Impact      string  `json:"impact"` // storage, organization, quality
	Confidence  float64 `json:"confidence"`
}

// calculateStatistics computes detailed statistics from the scan report
func (j *JSONReportGenerator) calculateStatistics(scanReport *api.ScanReport) *ReportStatistics {
	stats := &ReportStatistics{
		QualityDistribution: make(map[string]int),
		FormatDistribution:  make(map[string]int),
		SizeDistribution:    make(map[string]int),
	}

	// This would be populated from actual fingerprint data
	// For now, we'll use placeholder values
	stats.TotalSizeMB = 0 // Would be calculated from actual data
	stats.AverageFileSizeMB = 0
	stats.AverageQuality = 0

	return stats
}

// generateRecommendations generates actionable recommendations
func (j *JSONReportGenerator) generateRecommendations(scanReport *api.ScanReport) []*Recommendation {
	var recommendations []*Recommendation

	// Storage optimization recommendations
	if scanReport.ExactDuplicateCount > 0 {
		recommendations = append(recommendations, &Recommendation{
			Type:        "storage_optimization",
			Priority:    "high",
			Description: fmt.Sprintf("Found %d exact duplicate groups", scanReport.ExactDuplicateCount),
			Action:      "Run clean command to remove exact duplicates",
			Impact:      "storage",
			Confidence:  1.0,
		})
	}

	if scanReport.NearDuplicateCount > 0 {
		recommendations = append(recommendations, &Recommendation{
			Type:        "organization",
			Priority:    "medium",
			Description: fmt.Sprintf("Found %d near-duplicate groups", scanReport.NearDuplicateCount),
			Action:      "Review near-duplicates and keep best quality versions",
			Impact:      "organization",
			Confidence:  0.8,
		})
	}

	// Quality recommendations
	if scanReport.ProcessedImages > 100 {
		recommendations = append(recommendations, &Recommendation{
			Type:        "quality_audit",
			Priority:    "low",
			Description: "Large image collection detected",
			Action:      "Consider running quality analysis on all images",
			Impact:      "quality",
			Confidence:  0.6,
		})
	}

	return recommendations
}

// GenerateSummary generates a summary JSON report
func (j *JSONReportGenerator) GenerateSummary(scanReport *api.ScanReport, outputPath string) error {
	summary := map[string]interface{}{
		"scan_id":          scanReport.ScanID,
		"total_files":      scanReport.TotalFiles,
		"processed_images": scanReport.ProcessedImages,
		"exact_duplicates": scanReport.ExactDuplicateCount,
		"near_duplicates":  scanReport.NearDuplicateCount,
		"scan_duration":    scanReport.ScanDuration.String(),
		"started_at":       scanReport.StartedAt.Format(time.RFC3339),
		"completed_at":     scanReport.CompletedAt.Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write summary JSON report: %w", err)
	}

	j.logger.Infof("JSON summary report saved to: %s", outputPath)
	return nil
}
