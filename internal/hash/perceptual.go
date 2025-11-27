package hash

import (
	"fmt"
	"image"

	"github.com/HaiderBassem/imaged/internal/hash/perceptual"
)

// PerceptualHash provides a unified interface for all perceptual hash algorithms
type PerceptualHash struct {
	aHash *perceptual.AHash
	pHash *perceptual.PHash
	dHash *perceptual.DHash
	wHash *perceptual.WHash
}

// NewPerceptualHash creates a new perceptual hash calculator
func NewPerceptualHash() *PerceptualHash {
	return &PerceptualHash{
		aHash: perceptual.NewAHash(8),
		pHash: perceptual.NewPHash(32, 8),
		dHash: perceptual.NewDHash(9, 8),
		wHash: perceptual.NewWHash(64),
	}
}

// ComputeAll computes all perceptual hashes for an image
func (p *PerceptualHash) ComputeAll(img image.Image) (map[string]uint64, error) {
	hashes := make(map[string]uint64)

	// Compute AHash
	aHash, err := p.aHash.Compute(img)
	if err != nil {
		return nil, err
	}
	hashes["ahash"] = aHash

	// Compute PHash
	pHash, err := p.pHash.Compute(img)
	if err != nil {
		return nil, err
	}
	hashes["phash"] = pHash

	// Compute DHash
	dHash, err := p.dHash.Compute(img)
	if err != nil {
		return nil, err
	}
	hashes["dhash"] = dHash

	// Compute WHash
	wHash, err := p.wHash.Compute(img)
	if err != nil {
		return nil, err
	}
	hashes["whash"] = wHash

	return hashes, nil
}

// ComputeSpecific computes a specific type of perceptual hash
func (p *PerceptualHash) ComputeSpecific(img image.Image, hashType string) (uint64, error) {
	switch hashType {
	case "ahash":
		return p.aHash.Compute(img)
	case "phash":
		return p.pHash.Compute(img)
	case "dhash":
		return p.dHash.Compute(img)
	case "whash":
		return p.wHash.Compute(img)
	default:
		return 0, fmt.Errorf("unknown hash type: %s", hashType)
	}
}

// CompareHashes compares two sets of perceptual hashes
func (p *PerceptualHash) CompareHashes(hashes1, hashes2 map[string]uint64) map[string]float64 {
	similarities := make(map[string]float64)

	for hashType, hash1 := range hashes1 {
		hash2, exists := hashes2[hashType]
		if !exists {
			continue
		}

		var similarity float64
		switch hashType {
		case "ahash":
			similarity = p.aHash.Similarity(hash1, hash2)
		case "phash":
			similarity = p.pHash.Similarity(hash1, hash2)
		case "dhash":
			similarity = p.dHash.Similarity(hash1, hash2)
		case "whash":
			similarity = p.wHash.Similarity(hash1, hash2)
		}

		similarities[hashType] = similarity
	}

	return similarities
}

// GetSupportedHashTypes returns all supported perceptual hash types
func (p *PerceptualHash) GetSupportedHashTypes() []string {
	return []string{"ahash", "phash", "dhash", "whash"}
}
