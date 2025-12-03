package engine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/HaiderBassem/imaged/internal/index"
	"github.com/HaiderBassem/imaged/internal/quality"
	"github.com/HaiderBassem/imaged/internal/scanner"
	"github.com/HaiderBassem/imaged/internal/similarity"
	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
)

// Engine is the central coordinator for all image processing operations
type Engine struct {
	config     EngineConfig
	index      index.Store
	scanner    *scanner.Scanner
	quality    *quality.Analyzer
	similarity *similarity.Comparator
	logger     *logrus.Logger
}

// EngineConfig defines the configuration for the image processing engine
type EngineConfig struct {
	IndexPath     string
	NumWorkers    int
	UseGPU        bool
	LogLevel      string
	MaxMemoryMB   int
	HashConfig    HashConfig
	QualityConfig quality.Config
}

// HashConfig defines which perceptual hash algorithms to compute
type HashConfig struct {
	ComputeAHash bool
	ComputePHash bool
	ComputeDHash bool
	ComputeWHash bool
	HashSize     int
}

// ScanProgress represents real-time scan progress state
type ScanProgress struct {
	Current     int     // number of processed files
	Total       int     // total files to process
	Percentage  float64 // completion percentage (0-100)
	CurrentFile string  // currently processed file path
}

// NewEngine creates a new image processing engine with the specified configuration
func NewEngine(cfg EngineConfig) (*Engine, error) {
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Initialize the index storage backend
	store, err := index.NewBoltStore(cfg.IndexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create index store: %w", err)
	}

	// Initialize the folder scanner
	scanner := scanner.NewScanner(scanner.Config{
		NumWorkers:       cfg.NumWorkers,
		SupportedFormats: []string{".jpg", ".jpeg", ".png", ".webp", ".tiff", ".bmp"},
	})

	// Initialize the quality analyzer
	qualityAnalyzer := quality.NewAnalyzer(cfg.QualityConfig)

	// Initialize the similarity comparator
	comparator := similarity.NewComparator(similarity.ComparatorConfig{
		MinSimilarity: 0.8,
		UseFeatureVec: false,
	})

	return &Engine{
		config:     cfg,
		index:      store,
		scanner:    scanner,
		quality:    qualityAnalyzer,
		similarity: comparator,
		logger:     logger,
	}, nil
}

// ScanFolder recursively scans a folder and indexes all discovered images
func (e *Engine) ScanFolder(ctx context.Context, folderPath string, progress chan<- api.ScanProgress) error {
	e.logger.Infof("Starting scan of folder: %s", folderPath)

	startTime := time.Now()

	// Perform the initial folder scan to discover image files
	imagePaths, err := e.scanner.ScanFolder(ctx, folderPath)
	if err != nil {
		return fmt.Errorf("failed to scan folder: %w", err)
	}

	total := len(imagePaths)
	e.logger.Infof("Found %d images to process", total)

	// Process each image file with progress reporting
	processed := 0
	for _, path := range imagePaths {
		select {
		case <-ctx.Done():
			e.logger.Info("Scan operation cancelled by user")
			return ctx.Err()
		default:
			fingerprint, err := e.processImage(path)
			if err != nil {
				e.logger.Warnf("Failed to process image %s: %v", path, err)
				continue
			}

			// Persist the computed fingerprint to the index
			if err := e.index.SaveFingerprint(fingerprint); err != nil {
				e.logger.Warnf("Failed to save fingerprint for %s: %v", path, err)
				continue
			}

			processed++

			// Report progress to the caller if channel is provided
			if progress != nil {
				progress <- api.ScanProgress{
					Current:     processed,
					Total:       total,
					CurrentFile: path,
					Percentage:  float64(processed) / float64(total) * 100,
				}
			}
		}
	}

	duration := time.Since(startTime)
	e.logger.Infof("Scan completed. Processed %d images in %v", processed, duration)
	return nil
}

// processImage performs comprehensive analysis on a single image file
func (e *Engine) processImage(path string) (api.ImageFingerprint, error) {
	var fingerprint api.ImageFingerprint

	// Generate unique ID for this image
	fingerprint.ID = api.ImageID(generateImageID(path))
	fingerprint.CreatedAt = time.Now()

	// Load and decode the image with metadata
	img, metadata, err := e.loadImage(path)
	if err != nil {
		return fingerprint, fmt.Errorf("failed to load image %s: %w", path, err)
	}

	fingerprint.Metadata = metadata

	// Compute perceptual hashes based on configuration
	if e.config.HashConfig.ComputeAHash {
		fingerprint.PHashes.AHash, err = e.computeAHash(img)
		if err != nil {
			e.logger.Warnf("Failed to compute AHash for %s: %v", path, err)
		}
	}

	if e.config.HashConfig.ComputePHash {
		fingerprint.PHashes.PHash, err = e.computePHash(img)
		if err != nil {
			e.logger.Warnf("Failed to compute PHash for %s: %v", path, err)
		}
	}

	if e.config.HashConfig.ComputeDHash {
		fingerprint.PHashes.DHash, err = e.computeDHash(img)
		if err != nil {
			e.logger.Warnf("Failed to compute DHash for %s: %v", path, err)
		}
	}

	if e.config.HashConfig.ComputeWHash {
		fingerprint.PHashes.WHash, err = e.computeWHash(img)
		if err != nil {
			e.logger.Warnf("Failed to compute WHash for %s: %v", path, err)
		}
	}

	// Analyze image quality
	qualityScore, err := e.quality.Analyze(img)
	if err != nil {
		e.logger.Warnf("Failed to analyze quality for %s: %v", path, err)
		// Set default quality values if analysis fails
		fingerprint.Quality = api.ImageQuality{
			FinalScore: 50.0, // Default medium quality
		}
	} else {
		fingerprint.Quality = *qualityScore
	}

	e.logger.Debugf("Processed image %s: Quality=%.1f, Hashes=[A:%016x P:%016x]",
		path, fingerprint.Quality.FinalScore, fingerprint.PHashes.AHash, fingerprint.PHashes.PHash)

	return fingerprint, nil
}

// loadImage handles image loading, decoding, and basic metadata extraction
func (e *Engine) loadImage(path string) (image.Image, api.ImageMetadata, error) {
	var metadata api.ImageMetadata
	metadata.Path = path

	// Get file information
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to get file info: %w", err)
	}

	metadata.SizeBytes = fileInfo.Size()
	metadata.ModifiedAt = fileInfo.ModTime()

	// Compute SHA256 hash of the file
	sha256Hash, err := e.computeFileHash(path)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to compute file hash: %w", err)
	}
	metadata.SHA256 = sha256Hash

	// Open and decode the image file
	file, err := os.Open(path)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Decode image to get format and dimensions
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to decode image: %w", err)
	}

	metadata.Format = format
	bounds := img.Bounds()
	metadata.Width = bounds.Dx()
	metadata.Height = bounds.Dy()

	// Reset file for EXIF extraction (if needed)
	file.Seek(0, 0)

	// Extract EXIF metadata (simplified - would use proper EXIF library)
	exifInfo, err := e.extractEXIFMetadata(file, path)
	if err != nil {
		e.logger.Debugf("Failed to extract EXIF metadata from %s: %v", path, err)
	} else {
		metadata.EXIF = exifInfo
	}

	return img, metadata, nil
}

// computeFileHash calculates the SHA256 hash of a file
func (e *Engine) computeFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// extractEXIFMetadata extracts EXIF metadata from an image file
func (e *Engine) extractEXIFMetadata(file *os.File, path string) (*api.EXIFInfo, error) {
	// This is a simplified implementation
	// In a real implementation, you would use a proper EXIF library like goexif
	exifInfo := &api.EXIFInfo{
		HasGPS: false,
	}

	// Placeholder for actual EXIF extraction
	// For now, we'll just set some basic info based on file extension
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg":
		exifInfo.CameraModel = "Unknown JPEG Camera"
	case ".png":
		exifInfo.CameraModel = "Unknown PNG Source"
	}

	return exifInfo, nil
}

// computeAHash calculates the Average Hash for an image
func (e *Engine) computeAHash(img image.Image) (uint64, error) {
	// Resize image to 8x8 for hash computation
	resized := imaging.Resize(img, 8, 8, imaging.Lanczos)
	gray := imaging.Grayscale(resized)

	bounds := gray.Bounds()
	var sum uint64
	var pixels []uint64

	// Calculate average pixel value
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := gray.At(x, y).RGBA()
			// Convert to grayscale value (simplified)
			luminance := (uint64(r) + uint64(g) + uint64(b)) / 3
			sum += luminance
			pixels = append(pixels, luminance)
		}
	}

	average := sum / uint64(len(pixels))
	var hash uint64

	// Create hash bits based on pixel values compared to average
	for i, pixel := range pixels {
		if pixel > average {
			hash |= 1 << uint(i)
		}
	}

	return hash, nil
}

// computePHash calculates the Perception Hash for an image using DCT
func (e *Engine) computePHash(img image.Image) (uint64, error) {
	// Resize to 32x32 for better frequency analysis
	resized := imaging.Resize(img, 32, 32, imaging.Lanczos)
	gray := imaging.Grayscale(resized)

	bounds := gray.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		return 0, fmt.Errorf("unexpected image dimensions for pHash")
	}

	// Create 32x32 matrix of luminance values
	var matrix [32][32]float64
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			r, g, b, _ := gray.At(x, y).RGBA()
			luminance := (float64(r) + float64(g) + float64(b)) / 3.0 / 65535.0
			matrix[y][x] = luminance
		}
	}

	// Simplified DCT implementation (in real code, use proper DCT)
	// For now, we'll use a simplified approach similar to AHash but on 32x32
	var sum float64
	var values []float64

	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			sum += matrix[y][x]
			values = append(values, matrix[y][x])
		}
	}

	average := sum / float64(len(values))
	var hash uint64

	for i, value := range values {
		if value > average {
			hash |= 1 << uint(i%64) // Use modulo to fit in 64 bits
		}
	}

	return hash, nil
}

// computeDHash calculates the Difference Hash for an image
func (e *Engine) computeDHash(img image.Image) (uint64, error) {
	// Resize to 9x8 for difference calculation (8x8 differences)
	resized := imaging.Resize(img, 9, 8, imaging.Lanczos)
	gray := imaging.Grayscale(resized)

	bounds := gray.Bounds()
	var hash uint64
	bitPosition := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X-1; x++ {
			// Get current and next pixel
			r1, g1, b1, _ := gray.At(x, y).RGBA()
			r2, g2, b2, _ := gray.At(x+1, y).RGBA()

			luminance1 := (uint64(r1) + uint64(g1) + uint64(b1)) / 3
			luminance2 := (uint64(r2) + uint64(g2) + uint64(b2)) / 3

			// Set bit if next pixel is brighter
			if luminance2 > luminance1 {
				hash |= 1 << uint(bitPosition)
			}
			bitPosition++

			// Only need 64 bits total
			if bitPosition >= 64 {
				break
			}
		}
		if bitPosition >= 64 {
			break
		}
	}

	return hash, nil
}

// computeWHash calculates the Wavelet Hash for an image
func (e *Engine) computeWHash(img image.Image) (uint64, error) {
	// For now, implement a simplified version using multiple scales
	// In a real implementation, you would use proper wavelet transforms
	return e.computeAHash(img) // Fallback to AHash for now
}

// FindExactDuplicates identifies images with identical content using cryptographic hashes
func (e *Engine) FindExactDuplicates() ([]api.DuplicateGroup, error) {
	e.logger.Info("Searching for exact duplicates using SHA256 hashes")

	// Retrieve all fingerprints from the index
	fingerprints, err := e.index.GetAllFingerprints()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve fingerprints: %w", err)
	}

	// Group images by their SHA256 hash
	hashGroups := make(map[string][]api.ImageID)
	for _, fp := range fingerprints {
		hashGroups[fp.Metadata.SHA256] = append(hashGroups[fp.Metadata.SHA256], fp.ID)
	}

	// Create duplicate groups for hashes with multiple images
	var groups []api.DuplicateGroup
	groupCounter := 0

	for _, imageIDs := range hashGroups {
		if len(imageIDs) > 1 {
			mainImage := e.selectBestImage(imageIDs, fingerprints, api.PolicyHighestQuality)

			groups = append(groups, api.DuplicateGroup{
				GroupID:      fmt.Sprintf("exact_%d", groupCounter),
				MainImage:    mainImage,
				DuplicateIDs: e.removeElement(imageIDs, mainImage),
				Reason:       "exact",
				Confidence:   1.0,
			})
			groupCounter++
		}
	}

	e.logger.Infof("Found %d exact duplicate groups", len(groups))
	return groups, nil
}

// FindNearDuplicates identifies visually similar images using perceptual hashing
func (e *Engine) FindNearDuplicates(threshold float64) ([]api.DuplicateGroup, error) {
	e.logger.Infof("Searching for near duplicates with similarity threshold: %.2f", threshold)

	fingerprints, err := e.index.GetAllFingerprints()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve fingerprints: %w", err)
	}

	var groups []api.DuplicateGroup
	processed := make(map[api.ImageID]bool)
	groupCounter := 0

	for i, fp1 := range fingerprints {
		if processed[fp1.ID] {
			continue
		}

		var similarImages []api.ImageID
		similarImages = append(similarImages, fp1.ID)
		processed[fp1.ID] = true

		for j := i + 1; j < len(fingerprints); j++ {
			fp2 := fingerprints[j]
			if processed[fp2.ID] {
				continue
			}

			similarity, err := e.similarity.CompareFingerprints(fp1, fp2)
			if err != nil {
				e.logger.Warnf("Failed to compare %s and %s: %v", fp1.ID, fp2.ID, err)
				continue
			}

			if similarity >= threshold {
				similarImages = append(similarImages, fp2.ID)
				processed[fp2.ID] = true
			}
		}

		if len(similarImages) > 1 {
			mainImage := e.selectBestImage(similarImages, fingerprints, api.PolicyHighestQuality)

			groups = append(groups, api.DuplicateGroup{
				GroupID:      fmt.Sprintf("near_%d", groupCounter),
				MainImage:    mainImage,
				DuplicateIDs: e.removeElement(similarImages, mainImage),
				Reason:       "near",
				Confidence:   e.calculateGroupConfidence(similarImages, fingerprints),
			})
			groupCounter++
		}
	}

	e.logger.Infof("Found %d near-duplicate groups", len(groups))
	return groups, nil
}

// RateImageQuality analyzes and rates the quality of a specific image
func (e *Engine) RateImageQuality(imagePath string) (*api.ImageQuality, error) {
	e.logger.Debugf("Analyzing image quality: %s", imagePath)

	img, _, err := e.loadImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image for quality analysis: %w", err)
	}

	quality, err := e.quality.Analyze(img)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze image quality: %w", err)
	}

	e.logger.Debugf("Quality analysis completed for %s: Score=%.1f", imagePath, quality.FinalScore)
	return quality, nil
}

// CleanDuplicates performs duplicate cleaning based on the provided options
func (e *Engine) CleanDuplicates(options api.CleanOptions) (*api.CleanReport, error) {
	e.logger.Info("Starting duplicate cleaning process")

	report := &api.CleanReport{}
	startTime := time.Now()

	// 1) Find exact duplicates
	exactGroups, err := e.FindExactDuplicates()
	if err != nil {
		return nil, err
	}

	// 2) Find near duplicates (cropped, resized, zoomed...)
	nearGroups, err := e.FindNearDuplicates(options.MaxSimilarityThreshold)
	if err != nil {
		return nil, err
	}

	report.TotalProcessed = len(exactGroups) + len(nearGroups)

	movedExact, freedExact := e.processExactGroups(exactGroups, options)

	// process near-duplicate groups the same way
	movedNear, freedNear := e.processNearGroups(nearGroups, options)

	report.MovedFiles = movedExact + movedNear
	report.FreedSpace = freedExact + freedNear

	e.logger.Infof("Clean completed: %d files moved, %s freed in %v",
		report.MovedFiles,
		FormatBytes(report.FreedSpace),
		time.Since(startTime),
	)

	return report, nil
}

// processNearGroups handles the movement/deletion of near-duplicate files in a group
func (e *Engine) processNearGroups(groups []api.DuplicateGroup, options api.CleanOptions) (int, int64) {
	var moved int
	var freed int64

	for _, group := range groups {
		for _, dupID := range group.DuplicateIDs {
			fp, err := e.index.GetFingerprint(dupID)
			if err != nil {
				continue
			}

			// Skip if quality is below threshold
			if fp.Quality.FinalScore < options.MinQualityScore {
				e.logger.Debugf("Skipping low quality near-duplicate image: %s (score: %.1f)",
					fp.Metadata.Path, fp.Quality.FinalScore)
				continue
			}

			src := fp.Metadata.Path
			dstDir := filepath.Join(options.OutputDir, group.GroupID)
			_ = os.MkdirAll(dstDir, 0755)
			dst := filepath.Join(dstDir, filepath.Base(src))

			if options.DryRun {
				e.logger.Infof("DRY RUN: would move near-duplicate %s -> %s", src, dst)
				continue
			}

			if err := os.Rename(src, dst); err != nil {
				e.logger.Warnf("Failed to move near-duplicate %s: %v", src, err)
				continue
			}

			moved++
			freed += fp.Metadata.SizeBytes
		}
	}

	return moved, freed
}

func (e *Engine) verifyRealBinaryMatch(main api.ImageID, duplicates []api.ImageID) (bool, error) {
	mainFP, err := e.index.GetFingerprint(main)
	if err != nil {
		return false, err
	}

	mainBytes, err := os.ReadFile(mainFP.Metadata.Path)
	if err != nil {
		return false, err
	}

	for _, id := range duplicates {
		fp, err := e.index.GetFingerprint(id)
		if err != nil {
			return false, err
		}

		dupBytes, err := os.ReadFile(fp.Metadata.Path)
		if err != nil {
			return false, err
		}

		if !bytes.Equal(mainBytes, dupBytes) {
			return false, nil
		}
	}

	return true, nil
}

func (e *Engine) processExactGroups(groups []api.DuplicateGroup, options api.CleanOptions) (int, int64) {
	var moved int
	var freed int64

	for _, group := range groups {

		ok, err := e.verifyRealBinaryMatch(group.MainImage, group.DuplicateIDs)
		if err != nil || !ok {
			continue
		}

		for _, dupID := range group.DuplicateIDs {
			fp, err := e.index.GetFingerprint(dupID)
			if err != nil {
				continue
			}

			src := fp.Metadata.Path
			dstDir := filepath.Join(options.OutputDir, group.GroupID)
			_ = os.MkdirAll(dstDir, 0755)
			dst := filepath.Join(dstDir, filepath.Base(src))

			if options.DryRun {
				e.logger.Infof("DRY RUN: would move %s -> %s", src, dst)
				continue
			}

			if err := os.Rename(src, dst); err != nil {
				e.logger.Warnf("Failed to move %s: %v", src, err)
				continue
			}

			moved++
			freed += fp.Metadata.SizeBytes
		}
	}

	return moved, freed
}

// processDuplicateGroup handles the movement/deletion of duplicate files in a group
func (e *Engine) ProcessDuplicateGroup(group api.DuplicateGroup, options api.CleanOptions) (int, error) {
	moved := 0

	for _, duplicateID := range group.DuplicateIDs {
		fingerprint, err := e.index.GetFingerprint(duplicateID)
		if err != nil {
			return moved, fmt.Errorf("failed to get fingerprint for %s: %w", duplicateID, err)
		}

		// Skip if quality is below threshold
		if fingerprint.Quality.FinalScore < options.MinQualityScore {
			e.logger.Debugf("Skipping low quality image: %s (score: %.1f)",
				fingerprint.Metadata.Path, fingerprint.Quality.FinalScore)
			continue
		}

		if options.MoveDuplicates {
			err := e.MoveDuplicate(fingerprint, options.OutputDir, group.GroupID)
			if err != nil {
				return moved, fmt.Errorf("failed to move duplicate %s: %w", duplicateID, err)
			}
		} else {
			err := e.deleteDuplicate(fingerprint)
			if err != nil {
				return moved, fmt.Errorf("failed to delete duplicate %s: %w", duplicateID, err)
			}
		}

		moved++
	}

	return moved, nil
}

// moveDuplicate moves a duplicate file to the output directory
func (e *Engine) MoveDuplicate(fp *api.ImageFingerprint, outputDir string, groupID string) error {
	sourcePath := fp.Metadata.Path
	filename := filepath.Base(sourcePath)
	destPath := filepath.Join(outputDir, groupID, filename)

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Move the file
	if err := os.Rename(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	// Update the fingerprint path
	fp.Metadata.Path = destPath
	if err := e.index.SaveFingerprint(*fp); err != nil {
		e.logger.Warnf("Failed to update fingerprint after move: %v", err)
	}

	e.logger.Debugf("Moved duplicate: %s -> %s", sourcePath, destPath)
	return nil
}

// deleteDuplicate permanently deletes a duplicate file
func (e *Engine) deleteDuplicate(fp *api.ImageFingerprint) error {
	if err := os.Remove(fp.Metadata.Path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Remove from index
	if err := e.index.DeleteFingerprint(fp.ID); err != nil {
		e.logger.Warnf("Failed to remove fingerprint after deletion: %v", err)
	}

	e.logger.Debugf("Deleted duplicate: %s", fp.Metadata.Path)
	return nil
}

// selectBestImage chooses the best image from a set based on selection policy
func (e *Engine) selectBestImage(images []api.ImageID, fingerprints []api.ImageFingerprint, policy api.SelectionPolicy) api.ImageID {
	if len(images) == 0 {
		return ""
	}

	// Create a map for quick fingerprint lookup
	fpMap := make(map[api.ImageID]api.ImageFingerprint)
	for _, fp := range fingerprints {
		fpMap[fp.ID] = fp
	}

	bestImage := images[0]
	bestScore := e.calculateImageScore(fpMap[bestImage], policy)

	for _, imgID := range images[1:] {
		fp, exists := fpMap[imgID]
		if !exists {
			continue
		}
		score := e.calculateImageScore(fp, policy)
		if score > bestScore {
			bestImage = imgID
			bestScore = score
		}
	}

	return bestImage
}

// calculateImageScore computes a score for an image based on selection policy
func (e *Engine) calculateImageScore(fp api.ImageFingerprint, policy api.SelectionPolicy) float64 {
	switch policy {
	case api.PolicyHighestQuality:
		return fp.Quality.FinalScore
	case api.PolicyHighestResolution:
		return float64(fp.Metadata.Width * fp.Metadata.Height)
	case api.PolicyBestExposure:
		// Score based on how close exposure is to ideal (0.5)
		return 1.0 - math.Abs(fp.Quality.Exposure-0.5)*2
	case api.PolicyOldest:
		return -float64(fp.Metadata.ModifiedAt.Unix()) // Negative for oldest first
	case api.PolicyNewest:
		return float64(fp.Metadata.ModifiedAt.Unix())
	default:
		return fp.Quality.FinalScore
	}
}

// calculateGroupConfidence computes the confidence level for a duplicate group
func (e *Engine) calculateGroupConfidence(images []api.ImageID, fingerprints []api.ImageFingerprint) float64 {
	if len(images) < 2 {
		return 0.0
	}

	fpMap := make(map[api.ImageID]api.ImageFingerprint)
	for _, fp := range fingerprints {
		fpMap[fp.ID] = fp
	}

	var totalSimilarity float64
	var comparisons int

	// Compare each pair in the group
	for i := 0; i < len(images); i++ {
		for j := i + 1; j < len(images); j++ {
			fp1, exists1 := fpMap[images[i]]
			fp2, exists2 := fpMap[images[j]]

			if exists1 && exists2 {
				similarity, err := e.similarity.CompareFingerprints(fp1, fp2)
				if err == nil {
					totalSimilarity += similarity
					comparisons++
				}
			}
		}
	}

	if comparisons == 0 {
		return 0.0
	}

	return totalSimilarity / float64(comparisons)
}

// calculateFreedSpace estimates the amount of storage space freed by cleaning
func (e *Engine) calculateFreedSpace(movedFiles int) int64 {
	// This is a simplified calculation
	// In a real implementation, you would sum the actual file sizes
	return int64(movedFiles) * 5 * 1024 * 1024 // Estimate 5MB per file
}

// removeElement removes a specific element from a slice
func (e *Engine) removeElement(slice []api.ImageID, element api.ImageID) []api.ImageID {
	result := make([]api.ImageID, 0, len(slice)-1)
	for _, item := range slice {
		if item != element {
			result = append(result, item)
		}
	}
	return result
}

// GetStats returns statistics about the image index
func (e *Engine) GetStats() (*index.Stats, error) {
	return e.index.GetStats()
}

// Close safely closes the engine and releases all resources
func (e *Engine) Close() error {
	e.logger.Info("Closing image processing engine")

	if e.index != nil {
		return e.index.Close()
	}

	return nil
}

// generateImageID creates a unique identifier for an image
func generateImageID(path string) string {
	hash := sha256.Sum256([]byte(path + time.Now().String()))
	return "img_" + hex.EncodeToString(hash[:8])
}

// formatBytes converts byte count to human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
