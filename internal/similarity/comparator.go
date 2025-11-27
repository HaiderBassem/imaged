package similarity

import (
	"math"

	"github.com/HaiderBassem/imaged/pkg/api"
)

// Comparator handles image similarity comparison using multiple algorithms
type Comparator struct {
	config ComparatorConfig
}

// ComparatorConfig defines similarity comparison parameters
type ComparatorConfig struct {
	MinSimilarity float64
	UseFeatureVec bool
	AHashWeight   float64
	PHashWeight   float64
	DHashWeight   float64
	WHashWeight   float64
}

// NewComparator creates a new similarity comparator
func NewComparator(cfg ComparatorConfig) *Comparator {
	// Set default weights if not specified
	if cfg.AHashWeight+cfg.PHashWeight+cfg.DHashWeight+cfg.WHashWeight == 0 {
		cfg.AHashWeight = 0.2
		cfg.PHashWeight = 0.4
		cfg.DHashWeight = 0.3
		cfg.WHashWeight = 0.1
	}

	return &Comparator{
		config: cfg,
	}
}

// CompareFingerprints calculates similarity between two image fingerprints
func (c *Comparator) CompareFingerprints(fp1, fp2 api.ImageFingerprint) (float64, error) {
	var totalSimilarity float64
	var totalWeight float64

	// Compare each type of hash with its respective weight
	if c.config.AHashWeight > 0 && fp1.PHashes.AHash != 0 && fp2.PHashes.AHash != 0 {
		similarity := c.compareAHash(fp1.PHashes.AHash, fp2.PHashes.AHash)
		totalSimilarity += similarity * c.config.AHashWeight
		totalWeight += c.config.AHashWeight
	}

	if c.config.PHashWeight > 0 && fp1.PHashes.PHash != 0 && fp2.PHashes.PHash != 0 {
		similarity := c.comparePHash(fp1.PHashes.PHash, fp2.PHashes.PHash)
		totalSimilarity += similarity * c.config.PHashWeight
		totalWeight += c.config.PHashWeight
	}

	if c.config.DHashWeight > 0 && fp1.PHashes.DHash != 0 && fp2.PHashes.DHash != 0 {
		similarity := c.compareDHash(fp1.PHashes.DHash, fp2.PHashes.DHash)
		totalSimilarity += similarity * c.config.DHashWeight
		totalWeight += c.config.DHashWeight
	}

	if c.config.WHashWeight > 0 && fp1.PHashes.WHash != 0 && fp2.PHashes.WHash != 0 {
		similarity := c.compareWHash(fp1.PHashes.WHash, fp2.PHashes.WHash)
		totalSimilarity += similarity * c.config.WHashWeight
		totalWeight += c.config.WHashWeight
	}

	if totalWeight == 0 {
		return 0.0, nil
	}

	finalSimilarity := totalSimilarity / totalWeight
	return math.Max(0.0, math.Min(1.0, finalSimilarity)), nil
}

// compareAHash compares two Average Hashes
func (c *Comparator) compareAHash(hash1, hash2 uint64) float64 {
	distance := hammingDistance(hash1, hash2)
	maxDistance := 64.0 // 64 bits for all hash types
	similarity := 1.0 - (float64(distance) / maxDistance)
	return math.Max(0.0, similarity)
}

// comparePHash compares two Perception Hashes
func (c *Comparator) comparePHash(hash1, hash2 uint64) float64 {
	distance := hammingDistance(hash1, hash2)
	maxDistance := 64.0
	similarity := 1.0 - (float64(distance) / maxDistance)

	// pHash is more sensitive to small differences
	if similarity > 0.9 {
		return similarity // High confidence for very similar images
	} else if similarity > 0.7 {
		return similarity * 0.9 // Slightly penalize medium similarity
	} else {
		return similarity * 0.8 // Penalize low similarity more
	}
}

// compareDHash compares two Difference Hashes
func (c *Comparator) compareDHash(hash1, hash2 uint64) float64 {
	distance := hammingDistance(hash1, hash2)
	maxDistance := 64.0
	similarity := 1.0 - (float64(distance) / maxDistance)
	return similarity
}

// compareWHash compares two Wavelet Hashes
func (c *Comparator) compareWHash(hash1, hash2 uint64) float64 {
	distance := hammingDistance(hash1, hash2)
	maxDistance := 64.0
	similarity := 1.0 - (float64(distance) / maxDistance)

	// wHash is good for scaled/cropped images, so we're more lenient
	if similarity > 0.6 {
		return math.Min(1.0, similarity*1.1) // Boost similarity for decent matches
	}
	return similarity
}

// hammingDistance calculates the Hamming distance between two 64-bit integers
func hammingDistance(a, b uint64) int {
	xor := a ^ b
	distance := 0
	for xor != 0 {
		distance++
		xor &= xor - 1
	}
	return distance
}

// FindSimilarImages finds all images similar to a target fingerprint
func (c *Comparator) FindSimilarImages(target api.ImageFingerprint, candidates []api.ImageFingerprint, threshold float64) []api.ImageFingerprint {
	var similar []api.ImageFingerprint

	for _, candidate := range candidates {
		if candidate.ID == target.ID {
			continue // Skip the target itself
		}

		similarity, err := c.CompareFingerprints(target, candidate)
		if err != nil {
			continue
		}

		if similarity >= threshold {
			similar = append(similar, candidate)
		}
	}

	return similar
}

// ClusterImages groups similar images into clusters
func (c *Comparator) ClusterImages(fingerprints []api.ImageFingerprint, threshold float64) []api.Cluster {
	var clusters []api.Cluster
	assigned := make(map[api.ImageID]bool)

	for i, fp1 := range fingerprints {
		if assigned[fp1.ID] {
			continue
		}

		cluster := api.Cluster{
			ClusterID: generateClusterID(len(clusters)),
			Images:    []api.ImageID{fp1.ID},
		}
		assigned[fp1.ID] = true

		for j := i + 1; j < len(fingerprints); j++ {
			fp2 := fingerprints[j]
			if assigned[fp2.ID] {
				continue
			}

			similarity, err := c.CompareFingerprints(fp1, fp2)
			if err != nil {
				continue
			}

			if similarity >= threshold {
				cluster.Images = append(cluster.Images, fp2.ID)
				assigned[fp2.ID] = true
			}
		}

		if len(cluster.Images) > 0 {
			clusters = append(clusters, cluster)
		}
	}

	return clusters
}

// // generateClusterID creates a unique identifier for a cluster
// func generateClusterID() string {
// 	return fmt.Sprintf("cluster_%d", time.Now().UnixNano())
// }
