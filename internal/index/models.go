package index

import "github.com/HaiderBassem/imaged/pkg/api"

// Store defines the interface for index storage operations
type Store interface {
	SaveFingerprint(fp api.ImageFingerprint) error
	GetFingerprint(id api.ImageID) (*api.ImageFingerprint, error)
	GetAllFingerprints() ([]api.ImageFingerprint, error)
	FindBySHA256(hash string) ([]api.ImageFingerprint, error)
	FindSimilarHashes(targetHash uint64, maxDistance int, hashType string) ([]api.ImageFingerprint, error)
	DeleteFingerprint(id api.ImageID) error
	GetStats() (*Stats, error)
	Close() error
	Compact() error
}

// Stats contains index statistics
type Stats struct {
	TotalImages     int64   `json:"total_images"`
	TotalSizeBytes  int64   `json:"total_size_bytes"`
	IndexSizeBytes  int64   `json:"index_size_bytes"`
	AverageQuality  float64 `json:"average_quality"`
	DuplicateGroups int     `json:"duplicate_groups"`
}

// BatchOperation represents a batch of index operations
type BatchOperation struct {
	SaveFingerprints   []api.ImageFingerprint
	DeleteFingerprints []api.ImageID
}

// QueryOptions provides options for index queries
type QueryOptions struct {
	Limit          int
	Offset         int
	MinQuality     float64
	MaxQuality     float64
	MinWidth       int
	MinHeight      int
	Formats        []string
	SortBy         string
	SortDescending bool
}
