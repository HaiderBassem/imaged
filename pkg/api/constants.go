package api

// Constants used throughout the image processing library

const (
	// Version information
	VersionMajor  = 1
	VersionMinor  = 0
	VersionPatch  = 0
	VersionString = "1.0.0"

	// Default values
	DefaultSimilarityThreshold = 0.8
	DefaultQualityThreshold    = 50.0
	DefaultMaxFileSize         = 500 * 1024 * 1024 // 500MB
	DefaultNumWorkers          = 4

	// Hash sizes
	DefaultHashSize = 8
	LargeHashSize   = 16
	MaximumHashSize = 32

	// Quality thresholds
	HighQualityThreshold   = 80.0
	MediumQualityThreshold = 60.0
	LowQualityThreshold    = 30.0

	// File format constants
	FormatJPEG = "jpeg"
	FormatPNG  = "png"
	FormatWEBP = "webp"
	FormatTIFF = "tiff"
	FormatBMP  = "bmp"
	FormatGIF  = "gif"

	// Duplicate reasons
	ReasonExact      = "exact"
	ReasonNear       = "near"
	ReasonResized    = "resized"
	ReasonCompressed = "compressed"
	ReasonCropped    = "cropped"

	// Performance constants
	MaxBatchSize     = 1000
	DefaultCacheSize = 1000
)

// SupportedFormats returns all supported image formats
func SupportedFormats() []string {
	return []string{
		".jpg", ".jpeg", ".png", ".webp",
		".tiff", ".tif", ".bmp", ".gif",
	}
}

// QualityLevel represents different quality tiers
type QualityLevel int

const (
	QualityExcellent QualityLevel = iota
	QualityGood
	QualityAverage
	QualityPoor
	QualityVeryPoor
)

// String returns the string representation of quality level
func (q QualityLevel) String() string {
	switch q {
	case QualityExcellent:
		return "excellent"
	case QualityGood:
		return "good"
	case QualityAverage:
		return "average"
	case QualityPoor:
		return "poor"
	case QualityVeryPoor:
		return "very_poor"
	default:
		return "unknown"
	}
}

// GetQualityLevel returns the quality level for a given score
func GetQualityLevel(score float64) QualityLevel {
	switch {
	case score >= 80:
		return QualityExcellent
	case score >= 70:
		return QualityGood
	case score >= 50:
		return QualityAverage
	case score >= 30:
		return QualityPoor
	default:
		return QualityVeryPoor
	}
}
