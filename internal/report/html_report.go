package report

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
)

// HTMLReportGenerator generates HTML format reports
type HTMLReportGenerator struct {
	logger *logrus.Logger
}

// NewHTMLReportGenerator creates a new HTML report generator
func NewHTMLReportGenerator() *HTMLReportGenerator {
	return &HTMLReportGenerator{
		logger: logrus.New(),
	}
}

// Generate generates a comprehensive HTML report
func (h *HTMLReportGenerator) Generate(scanReport *api.ScanReport, outputPath string) error {
	// Prepare data for template
	data := HTMLReportData{
		ScanReport:  scanReport,
		GeneratedAt: time.Now(),
		Statistics:  h.calculateStatistics(scanReport),
	}

	// Parse and execute template
	tmpl := template.Must(template.New("report").Funcs(template.FuncMap{
		"formatBytes": humanize.Bytes,
		"formatTime":  formatTime,
		"percent":     percent,
		"mul":         func(a, b float64) float64 { return a * b },
	}).Parse(htmlTemplate))

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create HTML report: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	h.logger.Infof("HTML report saved to: %s", outputPath)
	return nil
}

// HTMLReportData contains data for HTML report generation
type HTMLReportData struct {
	*api.ScanReport
	GeneratedAt time.Time
	Statistics  *HTMLStatistics
}

// HTMLStatistics contains statistics for HTML report
type HTMLStatistics struct {
	TotalSizeMB       float64
	AverageFileSizeMB float64
	AverageQuality    float64
	SpaceSavingsMB    float64
}

// calculateStatistics computes statistics for HTML report
func (h *HTMLReportGenerator) calculateStatistics(scanReport *api.ScanReport) *HTMLStatistics {
	// These would be calculated from actual data in a real implementation
	return &HTMLStatistics{
		TotalSizeMB:       0,                                           // Would be calculated
		AverageFileSizeMB: 0,                                           // Would be calculated
		AverageQuality:    0,                                           // Would be calculated
		SpaceSavingsMB:    float64(scanReport.ExactDuplicateCount) * 5, // Estimate
	}
}

// Template functions
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func percent(value, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) / float64(total) * 100
}

// HTML template
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Image Deduplication Report</title>
    <style>
        :root {
            --primary-color: #3498db;
            --secondary-color: #2c3e50;
            --success-color: #27ae60;
            --warning-color: #f39c12;
            --danger-color: #e74c3c;
            --light-bg: #f8f9fa;
            --border-color: #dee2e6;
        }
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f5f5f5;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .header {
            background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
            color: white;
            padding: 2rem;
            border-radius: 10px;
            margin-bottom: 2rem;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        
        .header h1 {
            font-size: 2.5rem;
            margin-bottom: 0.5rem;
        }
        
        .header .subtitle {
            font-size: 1.1rem;
            opacity: 0.9;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }
        
        .stat-card {
            background: white;
            padding: 1.5rem;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            text-align: center;
            border-left: 4px solid var(--primary-color);
        }
        
        .stat-card.highlight {
            border-left-color: var(--success-color);
            background: linear-gradient(135deg, #fff, #f8fff9);
        }
        
        .stat-number {
            font-size: 2rem;
            font-weight: bold;
            color: var(--secondary-color);
            margin-bottom: 0.5rem;
        }
        
        .stat-label {
            color: #666;
            font-size: 0.9rem;
        }
        
        .section {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            margin-bottom: 2rem;
        }
        
        .section-title {
            font-size: 1.5rem;
            color: var(--secondary-color);
            margin-bottom: 1.5rem;
            padding-bottom: 0.5rem;
            border-bottom: 2px solid var(--light-bg);
        }
        
        .duplicate-group {
            background: var(--light-bg);
            padding: 1rem;
            border-radius: 6px;
            margin-bottom: 1rem;
            border-left: 4px solid var(--warning-color);
        }
        
        .duplicate-group.exact {
            border-left-color: var(--danger-color);
        }
        
        .group-header {
            display: flex;
            justify-content: between;
            align-items: center;
            margin-bottom: 0.5rem;
        }
        
        .group-type {
            padding: 0.25rem 0.75rem;
            border-radius: 20px;
            font-size: 0.8rem;
            font-weight: bold;
            text-transform: uppercase;
        }
        
        .type-exact {
            background: var(--danger-color);
            color: white;
        }
        
        .type-near {
            background: var(--warning-color);
            color: white;
        }
        
        .confidence {
            font-size: 0.9rem;
            color: #666;
        }
        
        .file-list {
            max-height: 200px;
            overflow-y: auto;
            background: white;
            padding: 0.5rem;
            border-radius: 4px;
            font-family: monospace;
            font-size: 0.9rem;
        }
        
        .recommendations {
            background: linear-gradient(135deg, #fff8e1, #fff);
            border-left: 4px solid var(--warning-color);
        }
        
        .recommendation {
            display: flex;
            align-items: start;
            margin-bottom: 1rem;
            padding: 1rem;
            background: white;
            border-radius: 6px;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }
        
        .rec-icon {
            font-size: 1.5rem;
            margin-right: 1rem;
        }
        
        .rec-content h4 {
            color: var(--secondary-color);
            margin-bottom: 0.5rem;
        }
        
        .rec-content p {
            color: #666;
            margin-bottom: 0.5rem;
        }
        
        .rec-action {
            color: var(--primary-color);
            font-weight: bold;
        }
        
        .footer {
            text-align: center;
            padding: 2rem;
            color: #666;
            font-size: 0.9rem;
        }
        
        @media (max-width: 768px) {
            .container {
                padding: 10px;
            }
            
            .header h1 {
                font-size: 2rem;
            }
            
            .stats-grid {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 Image Deduplication Report</h1>
            <div class="subtitle">
                Generated on {{.GeneratedAt | formatTime}} | Scan ID: {{.ScanID}}
            </div>
        </div>

        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-number">{{.TotalFiles}}</div>
                <div class="stat-label">Total Files Scanned</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{{.ProcessedImages}}</div>
                <div class="stat-label">Images Processed</div>
            </div>
            <div class="stat-card highlight">
                <div class="stat-number">{{.ExactDuplicateCount}}</div>
                <div class="stat-label">Exact Duplicate Groups</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{{.NearDuplicateCount}}</div>
                <div class="stat-label">Near-Duplicate Groups</div>
            </div>
        </div>

        {{if .Groups}}
        <div class="section">
            <h2 class="section-title">🔍 Duplicate Analysis</h2>
            {{range .Groups}}
            <div class="duplicate-group {{if eq .Reason "exact"}}exact{{end}}">
                <div class="group-header">
                    <span class="group-type {{if eq .Reason "exact"}}type-exact{{else}}type-near{{end}}">
                        {{.Reason}} duplicate
                    </span>
                    <span class="confidence">Confidence: {{printf "%.0f" (mul .Confidence 100)}}%}}%}}%</span>
                </div>
                <div class="group-info">
                    <strong>Main Image:</strong> {{.MainImage}}<br>
                    <strong>Duplicates Found:</strong> {{len .DuplicateIDs}} files
                </div>
                {{if .DuplicateIDs}}
                <details style="margin-top: 0.5rem;">
                    <summary>Show duplicate files</summary>
                    <div class="file-list">
                        {{range .DuplicateIDs}}
                        <div>{{.}}</div>
                        {{end}}
                    </div>
                </details>
                {{end}}
            </div>
            {{end}}
        </div>
        {{end}}

        <div class="section recommendations">
            <h2 class="section-title">💡 Recommendations</h2>
            
            {{if gt .ExactDuplicateCount 0}}
            <div class="recommendation">
                <div class="rec-icon">✅</div>
                <div class="rec-content">
                    <h4>Remove Exact Duplicates</h4>
                    <p>You have {{.ExactDuplicateCount}} groups of exact duplicates that are wasting storage space.</p>
                    <div class="rec-action">Action: Run cleanup command to remove duplicates</div>
                </div>
            </div>
            {{end}}

            {{if gt .NearDuplicateCount 0}}
            <div class="recommendation">
                <div class="rec-icon">🔍</div>
                <div class="rec-content">
                    <h4>Review Near-Duplicates</h4>
                    <p>You have {{.NearDuplicateCount}} groups of similar images that might need manual review.</p>
                    <div class="rec-action">Action: Use find-duplicates command to review</div>
                </div>
            </div>
            {{end}}

            <div class="recommendation">
                <div class="rec-icon">📁</div>
                <div class="rec-content">
                    <h4>Organize Your Library</h4>
                    <p>Consider organizing images into folders based on date, event, or content.</p>
                    <div class="rec-action">Action: Use clustering features to organize</div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">📈 Scan Information</h2>
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                <div>
                    <strong>Scan Duration:</strong> {{.ScanDuration}}<br>
                    <strong>Started:</strong> {{.StartedAt | formatTime}}<br>
                    <strong>Completed:</strong> {{.CompletedAt | formatTime}}
                </div>
                <div>
                    <strong>Files Skipped:</strong> {{.SkippedFiles}}<br>
                    <strong>Total Clusters:</strong> {{len .Clusters}}<br>
                    <strong>Generated:</strong> {{.GeneratedAt | formatTime}}
                </div>
            </div>
        </div>

        <div class="footer">
            <p>Generated by <strong>ImageD</strong> - Professional Image Management Tool</p>
            <p>Report generated on {{.GeneratedAt | formatTime}}</p>
        </div>
    </div>
</body>
</html>`
