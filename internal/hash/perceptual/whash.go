package perceptual

import (
	"image"
	"math"

	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
)

// WHash computes the Wavelet Hash for image similarity
type WHash struct {
	Size int
}

// NewWHash creates a new Wavelet Hash calculator
func NewWHash(size int) *WHash {
	return &WHash{
		Size: size,
	}
}

// Compute calculates the Wavelet Hash for an image
func (w *WHash) Compute(img image.Image) (uint64, error) {
	// Convert to grayscale
	gray := imaging.Grayscale(img)

	// Resize to power of 2 for wavelet transform
	size := w.nearestPowerOf2(w.Size)
	resized := resize.Resize(uint(size), uint(size), gray, resize.Lanczos3)

	// Apply simplified Haar wavelet transform
	wavelet := w.haarWavelet(resized)

	// Take top-left 8x8 coefficients
	var hash uint64
	bitPos := 0

	// Calculate mean of wavelet coefficients
	var sum float64
	var count int

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if x == 0 && y == 0 {
				continue // Skip approximation coefficient
			}
			sum += math.Abs(wavelet[y][x])
			count++
		}
	}

	if count == 0 {
		return 0, nil
	}

	mean := sum / float64(count)

	// Create hash based on wavelet coefficients
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if x == 0 && y == 0 {
				continue
			}
			if math.Abs(wavelet[y][x]) > mean {
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

// haarWavelet applies Haar wavelet transform (simplified)
func (w *WHash) haarWavelet(img image.Image) [][]float64 {
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

	// Simplified Haar wavelet implementation
	// This is a basic version - real implementation would be more complex
	for level := size; level > 1; level /= 2 {
		for y := 0; y < level; y++ {
			for x := 0; x < level/2; x++ {
				// Horizontal transform
				avg := (matrix[y][2*x] + matrix[y][2*x+1]) / 2
				diff := (matrix[y][2*x] - matrix[y][2*x+1]) / 2
				matrix[y][x] = avg
				matrix[y][x+level/2] = diff
			}
		}

		for y := 0; y < level/2; y++ {
			for x := 0; x < level; x++ {
				// Vertical transform
				avg := (matrix[2*y][x] + matrix[2*y+1][x]) / 2
				diff := (matrix[2*y][x] - matrix[2*y+1][x]) / 2
				matrix[y][x] = avg
				matrix[y+level/2][x] = diff
			}
		}
	}

	return matrix
}

// nearestPowerOf2 finds the nearest power of 2 for a given size
func (w *WHash) nearestPowerOf2(size int) int {
	return int(math.Pow(2, math.Round(math.Log2(float64(size)))))
}

// Distance calculates the Hamming distance between two WHash values
func (w *WHash) Distance(hash1, hash2 uint64) int {
	xor := hash1 ^ hash2
	distance := 0
	for xor != 0 {
		distance++
		xor &= xor - 1
	}
	return distance
}

// Similarity calculates the similarity percentage between two hashes
func (w *WHash) Similarity(hash1, hash2 uint64) float64 {
	distance := w.Distance(hash1, hash2)
	return 1.0 - float64(distance)/64.0
}
