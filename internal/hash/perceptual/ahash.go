package perceptual

import (
	"image"

	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
)

// AHash computes the Average Hash for image similarity
type AHash struct {
	Size int
}

// NewAHash creates a new Average Hash calculator
func NewAHash(size int) *AHash {
	return &AHash{
		Size: size,
	}
}

// Compute calculates the Average Hash for an image
func (a *AHash) Compute(img image.Image) (uint64, error) {
	// Resize image to target size
	resized := resize.Resize(uint(a.Size), uint(a.Size), img, resize.Lanczos3)
	gray := imaging.Grayscale(resized)

	// Calculate average pixel value
	var sum uint64
	var pixels []uint64

	bounds := gray.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := gray.At(x, y).RGBA()
			// Convert to grayscale luminance
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

// Distance calculates the Hamming distance between two AHash values
func (a *AHash) Distance(hash1, hash2 uint64) int {
	xor := hash1 ^ hash2
	distance := 0
	for xor != 0 {
		distance++
		xor &= xor - 1
	}
	return distance
}

// Similarity calculates the similarity percentage between two hashes
func (a *AHash) Similarity(hash1, hash2 uint64) float64 {
	distance := a.Distance(hash1, hash2)
	maxDistance := a.Size * a.Size
	return 1.0 - float64(distance)/float64(maxDistance)
}
