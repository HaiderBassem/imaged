package perceptual

import (
	"image"

	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
)

// DHash computes the Difference Hash for image similarity
type DHash struct {
	Width  int
	Height int
}

// NewDHash creates a new Difference Hash calculator
func NewDHash(width, height int) *DHash {
	return &DHash{
		Width:  width,
		Height: height,
	}
}

// Compute calculates the Difference Hash for an image
func (d *DHash) Compute(img image.Image) (uint64, error) {
	// Resize to width+1 x height to allow difference calculation
	resized := resize.Resize(uint(d.Width+1), uint(d.Height), img, resize.Lanczos3)
	gray := imaging.Grayscale(resized)

	bounds := gray.Bounds()
	var hash uint64
	bitPosition := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X-1; x++ {
			// Get current and next pixel
			r1, g1, b1, _ := gray.At(x, y).RGBA()
			r2, g2, b2, _ := gray.At(x+1, y).RGBA()

			// Calculate luminance for both pixels
			luminance1 := (uint64(r1) + uint64(g1) + uint64(b1)) / 3
			luminance2 := (uint64(r2) + uint64(g2) + uint64(b2)) / 3

			// Set bit if next pixel is brighter
			if luminance2 > luminance1 {
				hash |= 1 << uint(bitPosition)
			}
			bitPosition++

			// Only need 64 bits total
			if bitPosition >= 64 {
				return hash, nil
			}
		}
	}

	return hash, nil
}

// Distance calculates the Hamming distance between two DHash values
func (d *DHash) Distance(hash1, hash2 uint64) int {
	xor := hash1 ^ hash2
	distance := 0
	for xor != 0 {
		distance++
		xor &= xor - 1
	}
	return distance
}

// Similarity calculates the similarity percentage between two hashes
func (d *DHash) Similarity(hash1, hash2 uint64) float64 {
	distance := d.Distance(hash1, hash2)
	return 1.0 - float64(distance)/64.0
}
