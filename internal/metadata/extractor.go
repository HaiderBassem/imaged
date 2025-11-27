package metadata

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/sirupsen/logrus"
)

// Extractor handles comprehensive metadata extraction from images
type Extractor struct {
	exifReader *EXIFReader
	logger     *logrus.Logger
}

// NewExtractor creates a new metadata extractor
func NewExtractor() *Extractor {
	return &Extractor{
		exifReader: NewEXIFReader(),
		logger:     logrus.New(),
	}
}

// ExtractMetadata extracts comprehensive metadata from an image file
func (e *Extractor) ExtractMetadata(filePath string) (*api.ImageMetadata, error) {
	metadata := &api.ImageMetadata{
		Path: filePath,
	}

	// Get basic file information
	if err := e.extractFileInfo(metadata); err != nil {
		return nil, fmt.Errorf("failed to extract file info: %w", err)
	}

	// Extract EXIF metadata if available
	if e.isEXIFSupported(filePath) {
		exifInfo, err := e.exifReader.ExtractEXIF(filePath)
		if err != nil {
			e.logger.Debugf("Failed to extract EXIF from %s: %v", filePath, err)
		} else {
			metadata.EXIF = exifInfo
		}
	}

	// Extract format-specific metadata
	if err := e.extractFormatSpecificMetadata(metadata); err != nil {
		e.logger.Debugf("Failed to extract format-specific metadata: %v", err)
	}

	return metadata, nil
}

// extractFileInfo extracts basic file system metadata
func (e *Extractor) extractFileInfo(metadata *api.ImageMetadata) error {
	fileInfo, err := os.Stat(metadata.Path)
	if err != nil {
		return err
	}

	metadata.SizeBytes = fileInfo.Size()
	metadata.ModifiedAt = fileInfo.ModTime()

	// Determine file format from extension
	ext := strings.ToLower(filepath.Ext(metadata.Path))
	switch ext {
	case ".jpg", ".jpeg":
		metadata.Format = "jpeg"
	case ".png":
		metadata.Format = "png"
	case ".webp":
		metadata.Format = "webp"
	case ".tiff", ".tif":
		metadata.Format = "tiff"
	case ".bmp":
		metadata.Format = "bmp"
	case ".gif":
		metadata.Format = "gif"
	default:
		metadata.Format = strings.TrimPrefix(ext, ".")
	}

	return nil
}

// extractFormatSpecificMetadata extracts metadata specific to image formats
func (e *Extractor) extractFormatSpecificMetadata(metadata *api.ImageMetadata) error {
	// Open file to read basic image dimensions
	file, err := os.Open(metadata.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode image config to get dimensions without loading full image
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("failed to decode image config: %w", err)
	}

	metadata.Width = config.Width
	metadata.Height = config.Height

	// Update format if not already set
	if metadata.Format == "" {
		metadata.Format = format
	}

	return nil
}

// isEXIFSupported checks if a file format typically contains EXIF data
func (e *Extractor) isEXIFSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	supported := []string{".jpg", ".jpeg", ".tiff", ".tif", ".cr2", ".nef", ".arw"}

	for _, supportedExt := range supported {
		if ext == supportedExt {
			return true
		}
	}

	return false
}

// ValidateMetadata checks if extracted metadata is valid and complete
func (e *Extractor) ValidateMetadata(metadata *api.ImageMetadata) error {
	if metadata.Path == "" {
		return fmt.Errorf("missing file path")
	}

	if metadata.SizeBytes <= 0 {
		return fmt.Errorf("invalid file size: %d", metadata.SizeBytes)
	}

	if metadata.Width <= 0 || metadata.Height <= 0 {
		return fmt.Errorf("invalid image dimensions: %dx%d", metadata.Width, metadata.Height)
	}

	if metadata.Format == "" {
		return fmt.Errorf("unknown image format")
	}

	// Check if file modification time is reasonable (not in future)
	if metadata.ModifiedAt.After(time.Now().Add(24 * time.Hour)) {
		e.logger.Warnf("Suspicious modification time for %s: %v", metadata.Path, metadata.ModifiedAt)
	}

	return nil
}

// GetMetadataSummary returns a summary of the metadata
func (e *Extractor) GetMetadataSummary(metadata *api.ImageMetadata) map[string]interface{} {
	summary := make(map[string]interface{})

	summary["path"] = metadata.Path
	summary["size_bytes"] = metadata.SizeBytes
	summary["format"] = metadata.Format
	summary["dimensions"] = fmt.Sprintf("%dx%d", metadata.Width, metadata.Height)
	summary["modified"] = metadata.ModifiedAt.Format(time.RFC3339)

	if metadata.EXIF != nil {
		exifSummary := make(map[string]interface{})
		if metadata.EXIF.CameraModel != "" {
			exifSummary["camera"] = metadata.EXIF.CameraModel
		}
		if metadata.EXIF.LensModel != "" {
			exifSummary["lens"] = metadata.EXIF.LensModel
		}
		if metadata.EXIF.ISO > 0 {
			exifSummary["iso"] = metadata.EXIF.ISO
		}
		if metadata.EXIF.Exposure != "" {
			exifSummary["exposure"] = metadata.EXIF.Exposure
		}
		if metadata.EXIF.Aperture > 0 {
			exifSummary["aperture"] = metadata.EXIF.Aperture
		}
		if !metadata.EXIF.TakenAt.IsZero() {
			exifSummary["taken_at"] = metadata.EXIF.TakenAt.Format(time.RFC3339)
		}
		if metadata.EXIF.HasGPS {
			exifSummary["gps"] = fmt.Sprintf("%f,%f", metadata.EXIF.GPSLat, metadata.EXIF.GPSLon)
		}

		summary["exif"] = exifSummary
	}

	return summary
}
