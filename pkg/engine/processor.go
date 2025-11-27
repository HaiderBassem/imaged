package engine

import (
	"context"
	"sync"

	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/sirupsen/logrus"
)

// Processor handles concurrent image processing
type Processor struct {
	engine  *Engine
	workers int
	logger  *logrus.Logger
}

// NewProcessor creates a new image processor
func NewProcessor(engine *Engine, workers int) *Processor {
	return &Processor{
		engine:  engine,
		workers: workers,
		logger:  logrus.New(),
	}
}

// ProcessBatch processes a batch of images concurrently
func (p *Processor) ProcessBatch(ctx context.Context, imagePaths []string) ([]api.ImageFingerprint, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	jobs := make(chan string, len(imagePaths))
	results := make(chan *api.ImageFingerprint, len(imagePaths))
	errors := make(chan error, len(imagePaths))

	var fingerprints []api.ImageFingerprint

	// Start worker goroutines
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go p.worker(ctx, i, jobs, results, errors, &wg)
	}

	// Send jobs
	go func() {
		for _, path := range imagePaths {
			jobs <- path
		}
		close(jobs)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Process results and errors
	for result := range results {
		if result != nil {
			mu.Lock()
			fingerprints = append(fingerprints, *result)
			mu.Unlock()
		}
	}

	// Check for errors
	var processErrors []error
	for err := range errors {
		if err != nil {
			processErrors = append(processErrors, err)
		}
	}

	if len(processErrors) > 0 {
		p.logger.Warnf("Completed with %d errors", len(processErrors))
	}

	return fingerprints, nil
}

// worker processes images from the jobs channel
func (p *Processor) worker(ctx context.Context, id int, jobs <-chan string, results chan<- *api.ImageFingerprint, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for path := range jobs {
		select {
		case <-ctx.Done():
			p.logger.Debugf("Worker %d stopping due to context cancellation", id)
			return
		default:
			fingerprint, err := p.engine.processImage(path)
			if err != nil {
				p.logger.Warnf("Worker %d failed to process %s: %v", id, path, err)
				errors <- err
				results <- nil
			} else {
				results <- &fingerprint
			}
		}
	}
}

// ProcessWithProgress processes images with progress reporting
func (p *Processor) ProcessWithProgress(ctx context.Context, imagePaths []string, progress chan<- api.ScanProgress) ([]api.ImageFingerprint, error) {
	total := len(imagePaths)
	processed := 0

	// Process in smaller batches to allow progress updates
	batchSize := 100
	var allFingerprints []api.ImageFingerprint

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		batch := imagePaths[i:end]
		fingerprints, err := p.ProcessBatch(ctx, batch)
		if err != nil {
			return nil, err
		}

		allFingerprints = append(allFingerprints, fingerprints...)
		processed += len(fingerprints)

		// Report progress
		if progress != nil {
			progress <- api.ScanProgress{
				Current:     processed,
				Total:       total,
				CurrentFile: batch[len(batch)-1], // Last file in batch
				Percentage:  float64(processed) / float64(total) * 100,
			}
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return allFingerprints, ctx.Err()
		default:
		}
	}

	return allFingerprints, nil
}
