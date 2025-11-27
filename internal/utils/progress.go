package utils

import (
	"fmt"
	"time"
)

// ProgressTracker tracks operation progress and calculates ETA
type ProgressTracker struct {
	Total       int
	Current     int
	Description string
	StartTime   time.Time
	lastUpdate  time.Time
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(total int, description string) *ProgressTracker {
	return &ProgressTracker{
		Total:       total,
		Description: description,
		StartTime:   time.Now(),
		lastUpdate:  time.Now(),
	}
}

// Update increments progress and displays status
func (p *ProgressTracker) Update(increment int) {
	p.Current += increment
	now := time.Now()

	// Throttle updates to avoid excessive console output
	if now.Sub(p.lastUpdate) < 100*time.Millisecond && p.Current < p.Total {
		return
	}

	p.lastUpdate = now
	p.display()
}

// Set directly sets the current progress
func (p *ProgressTracker) Set(current int) {
	p.Current = current
	p.display()
}

// display shows the current progress with ETA
func (p *ProgressTracker) display() {
	percentage := float64(p.Current) / float64(p.Total) * 100
	elapsed := time.Since(p.StartTime)

	var eta time.Duration
	if p.Current > 0 {
		totalEstimate := time.Duration(float64(elapsed) * float64(p.Total) / float64(p.Current))
		eta = totalEstimate - elapsed
	}

	fmt.Printf("\r%s: %.1f%% | %d/%d | Elapsed: %v | ETA: %v",
		p.Description, percentage, p.Current, p.Total,
		elapsed.Round(time.Second), eta.Round(time.Second))
}

// Complete marks the progress as complete
func (p *ProgressTracker) Complete() {
	p.Current = p.Total
	elapsed := time.Since(p.StartTime)
	fmt.Printf("\r%s: Complete! | Time: %v\n", p.Description, elapsed.Round(time.Second))
}

// ProgressBar creates a visual progress bar
func (p *ProgressTracker) ProgressBar(width int) string {
	if p.Total == 0 {
		return ""
	}

	progress := float64(p.Current) / float64(p.Total)
	bars := int(progress * float64(width))

	bar := "["
	for i := 0; i < width; i++ {
		if i < bars {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"

	return fmt.Sprintf("%s %.1f%%", bar, progress*100)
}
