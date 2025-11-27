package scanner

import (
	"context"
	"sync"
)

// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
	numWorkers int
	jobs       chan Job
	results    chan Result
	errors     chan error
	wg         sync.WaitGroup
}

// Job represents a unit of work for a worker
type Job struct {
	ID   int
	Path string
	Type JobType
}

// JobType defines the type of job
type JobType int

const (
	JobTypeScanDir JobType = iota
	JobTypeProcessFile
)

// Result represents the result of a job
type Result struct {
	JobID int
	Path  string
	Files []string
	Error error
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		jobs:       make(chan Job, numWorkers*2),
		results:    make(chan Result, numWorkers*2),
		errors:     make(chan error, numWorkers*2),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(ctx context.Context, scanner *Scanner) {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i, scanner)
	}
}

// worker processes jobs from the jobs channel
func (wp *WorkerPool) worker(ctx context.Context, id int, scanner *Scanner) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-wp.jobs:
			if !ok {
				return
			}
			wp.processJob(ctx, id, job, scanner)
		}
	}
}

// processJob processes a single job
func (wp *WorkerPool) processJob(ctx context.Context, workerID int, job Job, scanner *Scanner) {
	var result Result
	result.JobID = job.ID
	result.Path = job.Path

	switch job.Type {
	case JobTypeScanDir:
		files, err := scanner.scanDirectory(job.Path)
		if err != nil {
			result.Error = err
		} else {
			result.Files = files
		}
	case JobTypeProcessFile:
		// For file processing jobs
		result.Files = []string{job.Path}
	}

	select {
	case <-ctx.Done():
		return
	case wp.results <- result:
	}
}

// SubmitJob submits a job to the worker pool
func (wp *WorkerPool) SubmitJob(job Job) {
	wp.jobs <- job
}

// GetResults returns the results channel
func (wp *WorkerPool) GetResults() <-chan Result {
	return wp.results
}

// GetErrors returns the errors channel
func (wp *WorkerPool) GetErrors() <-chan error {
	return wp.errors
}

// Close closes the worker pool and waits for all workers to finish
func (wp *WorkerPool) Close() {
	close(wp.jobs)
	wp.wg.Wait()
	close(wp.results)
	close(wp.errors)
}

// ProcessBatch processes a batch of paths using the worker pool
func (wp *WorkerPool) ProcessBatch(ctx context.Context, paths []string, jobType JobType) []Result {
	var results []Result

	// Submit all jobs
	for i, path := range paths {
		job := Job{
			ID:   i,
			Path: path,
			Type: jobType,
		}
		wp.SubmitJob(job)
	}

	// Collect results
	for i := 0; i < len(paths); i++ {
		select {
		case <-ctx.Done():
			return results
		case result := <-wp.results:
			results = append(results, result)
		case err := <-wp.errors:
			// Handle errors
			_ = err
		}
	}

	return results
}

// GetStats returns worker pool statistics
func (wp *WorkerPool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"num_workers":       wp.numWorkers,
		"jobs_queue_len":    len(wp.jobs),
		"results_queue_len": len(wp.results),
	}
}
