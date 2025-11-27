package perceptual

import (
	"image"
	"math"

	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
)

// PHash computes the Perception Hash using DCT
type PHash struct {
	Size      int
	SmallSize int
}

// NewPHash creates a new Perception Hash calculator
func NewPHash(size, smallSize int) *PHash {
	return &PHash{
		Size:      size,
		SmallSize: smallSize,
	}
}

// Compute calculates the Perception Hash for an image
func (p *PHash) Compute(img image.Image) (uint64, error) {
	// Convert to grayscale and resize
	gray := imaging.Grayscale(img)
	resized := resize.Resize(uint(p.Size), uint(p.Size), gray, resize.Lanczos3)

	// Apply DCT (simplified implementation)
	dctMatrix := p.applyDCT(resized)

	// Take top-left 8x8 of DCT matrix (low frequencies)
	var hash uint64
	bitPos := 0

	// Calculate mean of DCT coefficients (excluding DC component)
	var sum float64
	var count int

	for y := 0; y < p.SmallSize; y++ {
		for x := 0; x < p.SmallSize; x++ {
			if x == 0 && y == 0 {
				continue // Skip DC component
			}
			sum += dctMatrix[y][x]
			count++
		}
	}

	if count == 0 {
		return 0, nil
	}

	mean := sum / float64(count)

	// Create hash based on DCT coefficients
	for y := 0; y < p.SmallSize; y++ {
		for x := 0; x < p.SmallSize; x++ {
			if x == 0 && y == 0 {
				continue // Skip DC component
			}
			if dctMatrix[y][x] > mean {
				hash |= 1 << uint(bitPos)
			}
			bitPos++
			if bitPos >= 64 {
				break
			}
		}
		if bitPos >= 64 {
			break
		}
	}

	return hash, nil
}

// applyDCT applies Discrete Cosine Transform to the image
func (p *PHash) applyDCT(img image.Image) [][]float64 {
	bounds := img.Bounds()
	size := bounds.Dx()

	// Create matrix from image pixels
	matrix := make([][]float64, size)
	for i := range matrix {
		matrix[i] = make([]float64, size)
	}

	// Fill matrix with pixel values
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			r, _, _, _ := img.At(x, y).RGBA()
			matrix[y][x] = float64(r) / 65535.0
		}
	}

	// Simplified DCT implementation
	// In production, use a proper DCT library
	dct := make([][]float64, size)
	for i := range dct {
		dct[i] = make([]float64, size)
	}

	// This is a simplified version - real DCT would be more complex
	for u := 0; u < size; u++ {
		for v := 0; v < size; v++ {
			var sum float64
			for i := 0; i < size; i++ {
				for j := 0; j < size; j++ {
					cos1 := math.Cos(float64((2*i+1)*u) * math.Pi / (2 * float64(size)))
					cos2 := math.Cos(float64((2*j+1)*v) * math.Pi / (2 * float64(size)))
					sum += matrix[i][j] * cos1 * cos2
				}
			}
			dct[u][v] = sum
		}
	}

	return dct
}

// Distance calculates the Hamming distance between two PHash values
func (p *PHash) Distance(hash1, hash2 uint64) int {
	xor := hash1 ^ hash2
	distance := 0
	for xor != 0 {
		distance++
		xor &= xor - 1
	}
	return distance
}

// Similarity calculates the similarity percentage between two hashes
func (p *PHash) Similarity(hash1, hash2 uint64) float64 {
	distance := p.Distance(hash1, hash2)
	return 1.0 - float64(distance)/64.0
}
