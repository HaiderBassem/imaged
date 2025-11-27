package index

import (
	"fmt"

	"github.com/HaiderBassem/imaged/pkg/api"
)

// StoreType represents the type of storage backend
type StoreType int

const (
	StoreTypeBoltDB StoreType = iota
	StoreTypeSQLite
	StoreTypeMemory
)

// Config defines index storage configuration
type Config struct {
	Type     StoreType
	Path     string
	ReadOnly bool
}

// NewStore creates a new index store based on configuration
func NewStore(cfg Config) (Store, error) {
	switch cfg.Type {
	case StoreTypeBoltDB:
		return NewBoltStore(cfg.Path)
	case StoreTypeSQLite:
		return NewSQLiteStore(cfg.Path)
	case StoreTypeMemory:
		return NewMemoryStore()
	default:
		return nil, fmt.Errorf("unsupported store type: %v", cfg.Type)
	}
}

// MemoryStore is an in-memory implementation for testing
type MemoryStore struct {
	fingerprints map[api.ImageID]api.ImageFingerprint
	sha256Index  map[string]api.ImageID
	pathIndex    map[string]api.ImageID
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() (*MemoryStore, error) {
	return &MemoryStore{
		fingerprints: make(map[api.ImageID]api.ImageFingerprint),
		sha256Index:  make(map[string]api.ImageID),
		pathIndex:    make(map[string]api.ImageID),
	}, nil
}

// SaveFingerprint stores a fingerprint in memory
func (m *MemoryStore) SaveFingerprint(fp api.ImageFingerprint) error {
	m.fingerprints[fp.ID] = fp
	m.sha256Index[fp.Metadata.SHA256] = fp.ID
	m.pathIndex[fp.Metadata.Path] = fp.ID
	return nil
}

// GetFingerprint retrieves a fingerprint from memory
func (m *MemoryStore) GetFingerprint(imageID api.ImageID) (*api.ImageFingerprint, error) {
	fp, exists := m.fingerprints[imageID]
	if !exists {
		return nil, api.ErrImageNotFound
	}
	return &fp, nil
}

// GetAllFingerprints returns all fingerprints from memory
func (m *MemoryStore) GetAllFingerprints() ([]api.ImageFingerprint, error) {
	fingerprints := make([]api.ImageFingerprint, 0, len(m.fingerprints))
	for _, fp := range m.fingerprints {
		fingerprints = append(fingerprints, fp)
	}
	return fingerprints, nil
}

// FindBySHA256 finds image by SHA256 hash in memory
func (m *MemoryStore) FindBySHA256(hash string) ([]api.ImageFingerprint, error) {
	imageID, exists := m.sha256Index[hash]
	if !exists {
		return []api.ImageFingerprint{}, nil
	}

	fp, exists := m.fingerprints[imageID]
	if !exists {
		return []api.ImageFingerprint{}, nil
	}

	return []api.ImageFingerprint{fp}, nil
}

// FindSimilarHashes placeholder for memory store
func (m *MemoryStore) FindSimilarHashes(targetHash uint64, maxDistance int, hashType string) ([]api.ImageFingerprint, error) {
	// Simple implementation that checks all fingerprints
	var similar []api.ImageFingerprint
	for _, fp := range m.fingerprints {
		var hashValue uint64
		switch hashType {
		case "ahash":
			hashValue = fp.PHashes.AHash
		case "phash":
			hashValue = fp.PHashes.PHash
		case "dhash":
			hashValue = fp.PHashes.DHash
		case "whash":
			hashValue = fp.PHashes.WHash
		default:
			continue
		}

		if hashValue == 0 {
			continue
		}

		distance := hammingDistance(targetHash, hashValue)
		if distance <= maxDistance {
			similar = append(similar, fp)
		}
	}
	return similar, nil
}

// DeleteFingerprint removes a fingerprint from memory
func (m *MemoryStore) DeleteFingerprint(imageID api.ImageID) error {
	fp, exists := m.fingerprints[imageID]
	if !exists {
		return api.ErrImageNotFound
	}

	delete(m.fingerprints, imageID)
	delete(m.sha256Index, fp.Metadata.SHA256)
	delete(m.pathIndex, fp.Metadata.Path)

	return nil
}

// GetStats returns memory store statistics
func (m *MemoryStore) GetStats() (*Stats, error) {
	var totalSize int64
	var totalQuality float64

	for _, fp := range m.fingerprints {
		totalSize += fp.Metadata.SizeBytes
		totalQuality += fp.Quality.FinalScore
	}

	count := int64(len(m.fingerprints))
	avgQuality := 0.0
	if count > 0 {
		avgQuality = totalQuality / float64(count)
	}

	return &Stats{
		TotalImages:    count,
		TotalSizeBytes: totalSize,
		AverageQuality: avgQuality,
	}, nil
}

// Close cleans up memory store
func (m *MemoryStore) Close() error {
	m.fingerprints = nil
	m.sha256Index = nil
	m.pathIndex = nil
	return nil
}

// Compact is a no-op for memory store
func (m *MemoryStore) Compact() error {
	return nil
}
