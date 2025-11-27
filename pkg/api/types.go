package api

import (
	"time"
)

// ImageID represents a unique identifier for an image
type ImageID string

// ImageMetadata contains comprehensive metadata about an image file
type ImageMetadata struct {
	Path       string    `json:"path"`
	SizeBytes  int64     `json:"size_bytes"`
	Format     string    `json:"format"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	ModifiedAt time.Time `json:"modified_at"`
	EXIF       *EXIFInfo `json:"exif,omitempty"`
	SHA256     string    `json:"sha256"`
}

// EXIFInfo contains EXIF metadata extracted from images
type EXIFInfo struct {
	CameraModel string    `json:"camera_model,omitempty"`
	LensModel   string    `json:"lens_model,omitempty"`
	ISO         int       `json:"iso,omitempty"`
	Exposure    string    `json:"exposure,omitempty"`
	Aperture    float64   `json:"aperture,omitempty"`
	FocalLength float64   `json:"focal_length,omitempty"`
	TakenAt     time.Time `json:"taken_at,omitempty"`
	GPSLat      float64   `json:"gps_lat,omitempty"`
	GPSLon      float64   `json:"gps_lon,omitempty"`
	HasGPS      bool      `json:"has_gps"`
}

// PerceptualHashes stores multiple perceptual hash values for an image
type PerceptualHashes struct {
	AHash uint64 `json:"a_hash"` // Average Hash - fast and good for exact matches
	PHash uint64 `json:"p_hash"` // Perception Hash - resistant to scaling and minor modifications
	DHash uint64 `json:"d_hash"` // Difference Hash - good for similar images
	WHash uint64 `json:"w_hash"` // Wavelet Hash - excellent for cropped/scaled images
}

// ImageQuality represents comprehensive quality analysis results
type ImageQuality struct {
	Sharpness   float64 `json:"sharpness"`   // 0..1 (1 = sharpest)
	Noise       float64 `json:"noise"`       // 0..1 (1 = most noisy)
	Exposure    float64 `json:"exposure"`    // 0..1 (0.5 = ideal exposure)
	Contrast    float64 `json:"contrast"`    // 0..1 (1 = best contrast)
	Compression float64 `json:"compression"` // 0..1 (1 = most artifacts)
	ColorCast   float64 `json:"color_cast"`  // 0..1 (1 = strongest color cast)
	FinalScore  float64 `json:"final_score"` // 0..100 overall quality score
}

// ImageFingerprint represents a complete digital fingerprint of an image
type ImageFingerprint struct {
	ID         ImageID          `json:"id"`
	Metadata   ImageMetadata    `json:"metadata"`
	PHashes    PerceptualHashes `json:"perceptual_hashes"`
	Quality    ImageQuality     `json:"quality"`
	ColorHist  []float64        `json:"color_histogram,omitempty"`
	FeatureVec []float32        `json:"feature_vector,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
}

// DuplicateGroup represents a group of duplicate or near-duplicate images
type DuplicateGroup struct {
	GroupID      string    `json:"group_id"`
	MainImage    ImageID   `json:"main_image"`
	DuplicateIDs []ImageID `json:"duplicate_ids"`
	Reason       string    `json:"reason"` // exact, near, resized, etc.
	Confidence   float64   `json:"confidence"`
}

// Cluster represents a group of similar images based on content analysis
type Cluster struct {
	ClusterID string    `json:"cluster_id"`
	Name      string    `json:"name,omitempty"`
	Images    []ImageID `json:"images"`
	Centroid  []float32 `json:"centroid,omitempty"`
}

// ScanProgress provides real-time progress information during scanning operations
type ScanProgress struct {
	Current     int     `json:"current"`
	Total       int     `json:"total"`
	CurrentFile string  `json:"current_file"`
	Percentage  float64 `json:"percentage"`
}

// ScanReport provides comprehensive results of a scanning operation
type ScanReport struct {
	ScanID              string           `json:"scan_id"`
	TotalFiles          int              `json:"total_files"`
	ProcessedImages     int              `json:"processed_images"`
	SkippedFiles        int              `json:"skipped_files"`
	ExactDuplicateCount int              `json:"exact_duplicate_count"`
	NearDuplicateCount  int              `json:"near_duplicate_count"`
	Groups              []DuplicateGroup `json:"duplicate_groups"`
	Clusters            []Cluster        `json:"clusters"`
	ScanDuration        time.Duration    `json:"scan_duration"`
	StartedAt           time.Time        `json:"started_at"`
	CompletedAt         time.Time        `json:"completed_at"`
}

// CleanOptions configures the behavior of duplicate cleaning operations
type CleanOptions struct {
	DryRun                 bool            `json:"dry_run"`
	SelectionPolicy        SelectionPolicy `json:"selection_policy"`
	MinQualityScore        float64         `json:"min_quality_score"`
	MaxSimilarityThreshold float64         `json:"max_similarity_threshold"`
	MoveDuplicates         bool            `json:"move_duplicates"`
	OutputDir              string          `json:"output_dir"`
}

// CleanReport provides results of a cleaning operation
type CleanReport struct {
	TotalProcessed int   `json:"total_processed"`
	MovedFiles     int   `json:"moved_files"`
	FreedSpace     int64 `json:"freed_space_bytes"`
	Errors         int   `json:"errors"`
}

// SelectionPolicy defines the strategy for selecting the best image from duplicates
type SelectionPolicy int

const (
	PolicyHighestQuality SelectionPolicy = iota
	PolicyHighestResolution
	PolicyBestExposure
	PolicyOldest
	PolicyNewest
)
