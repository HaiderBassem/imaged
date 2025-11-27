package quality

import (
	"image"
	"math"

	"github.com/disintegration/imaging"
)

// CompressionAnalyzer detects compression artifacts in images
type CompressionAnalyzer struct{}

// NewCompressionAnalyzer creates a new compression artifact analyzer
func NewCompressionAnalyzer() *CompressionAnalyzer {
	return &CompressionAnalyzer{}
}

// AnalyzeCompression detects JPEG compression artifacts
func (c *CompressionAnalyzer) AnalyzeCompression(img image.Image) (float64, error) {
	bounds := img.Bounds()
	if bounds.Dx() < 16 || bounds.Dy() < 16 {
		return 0.1, nil // Small images have less noticeable compression
	}

	// Analyze blockiness in JPEG images (common artifact)
	blockiness := c.analyzeBlockiness(img)

	// Analyze ringing artifacts (common in high compression)
	ringing := c.analyzeRingingArtifacts(img)

	// Analyze noise patterns that indicate compression
	noisePattern := c.analyzeCompressionNoise(img)

	// Combine artifacts into overall compression score
	compressionScore := (blockiness + ringing + noisePattern) / 3.0
	return math.Min(compressionScore, 1.0), nil
}

// analyzeBlockiness detects block artifacts from JPEG compression
func (c *CompressionAnalyzer) analyzeBlockiness(img image.Image) float64 {
	bounds := img.Bounds()
	gray := imaging.Grayscale(img)

	var blockArtifacts float64
	blockSize := 8 // JPEG uses 8x8 blocks

	// Check for discontinuities at block boundaries
	for y := blockSize; y < bounds.Dy()-blockSize; y += blockSize {
		for x := blockSize; x < bounds.Dx()-blockSize; x += blockSize {
			// Check horizontal block boundaries
			lr, _, _, _ := gray.At(x-1, y).RGBA()
			rr, _, _, _ := gray.At(x, y).RGBA()
			horizontalDiff := math.Abs(float64(lr) - float64(rr))

			// Check vertical block boundaries
			tr, _, _, _ := gray.At(x, y-1).RGBA()
			br, _, _, _ := gray.At(x, y).RGBA()
			verticalDiff := math.Abs(float64(tr) - float64(br))

			// High differences at block boundaries indicate compression artifacts
			if horizontalDiff > 10000 || verticalDiff > 10000 {
				blockArtifacts += 1.0
			}
		}
	}

	totalBlocks := ((bounds.Dx() / blockSize) - 1) * ((bounds.Dy() / blockSize) - 1)
	if totalBlocks == 0 {
		return 0.0
	}

	return math.Min(blockArtifacts/float64(totalBlocks), 1.0)
}

// analyzeRingingArtifacts detects ringing artifacts near edges
func (c *CompressionAnalyzer) analyzeRingingArtifacts(img image.Image) float64 {
	bounds := img.Bounds()
	gray := imaging.Grayscale(img)

	var ringingArtifacts float64
	edgeThreshold := 15000.0

	for y := 1; y < bounds.Dy()-1; y += 2 {
		for x := 1; x < bounds.Dx()-1; x += 2 {
			cr, _, _, _ := gray.At(x, y).RGBA()
			lr, _, _, _ := gray.At(x-1, y).RGBA()
			rr, _, _, _ := gray.At(x+1, y).RGBA()

			center := cr
			left := lr
			right := rr

			// Detect strong edges
			edgeStrength := math.Max(
				math.Abs(float64(center)-float64(left)),
				math.Abs(float64(center)-float64(right)),
			)

			if edgeStrength > edgeThreshold {
				// Check for oscillations near the edge (ringing)
				if x > 2 && x < bounds.Dx()-3 {
					leftOscillation := c.checkOscillation(gray, x-3, x, y)
					rightOscillation := c.checkOscillation(gray, x, x+3, y)

					if leftOscillation > 0.5 || rightOscillation > 0.5 {
						ringingArtifacts += 1.0
					}
				}
			}
		}
	}

	totalSamples := ((bounds.Dx() - 2) / 2) * ((bounds.Dy() - 2) / 2)
	if totalSamples == 0 {
		return 0.0
	}

	return math.Min(ringingArtifacts/float64(totalSamples), 1.0)
}

// checkOscillation checks for value oscillations indicating ringing
func (c *CompressionAnalyzer) checkOscillation(gray image.Image, startX, endX, y int) float64 {
	var oscillations int
	var lastDirection int

	pr, _, _, _ := gray.At(startX, y).RGBA()
	prevVal := pr

	for x := startX + 1; x <= endX; x++ {
		cr, _, _, _ := gray.At(x, y).RGBA()
		currentVal := cr

		diff := float64(currentVal) - float64(prevVal)

		currentDirection := 0
		if diff > 1000 {
			currentDirection = 1
		} else if diff < -1000 {
			currentDirection = -1
		}

		if lastDirection != 0 && currentDirection != 0 && currentDirection != lastDirection {
			oscillations++
		}

		lastDirection = currentDirection
		prevVal = currentVal
	}

	return float64(oscillations) / float64(endX-startX)
}

// analyzeCompressionNoise detects noise patterns characteristic of compression
func (c *CompressionAnalyzer) analyzeCompressionNoise(img image.Image) float64 {
	bounds := img.Bounds()
	grayImg := imaging.Grayscale(img)

	// Convert NRGBA to Grayscale image.Gray
	gray := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, grayImg.At(x, y))
		}
	}

	var highFreqNoise float64
	totalSamples := 0

	// Analyze high-frequency noise in smooth areas
	for y := 2; y < bounds.Dy()-2; y += 4 {
		for x := 2; x < bounds.Dx()-2; x += 4 {
			// Check if this is a smooth area
			if c.isSmoothArea(gray, x, y, 3) {
				noiseLevel := c.measureHighFrequencyNoise(gray, x, y)
				highFreqNoise += noiseLevel
				totalSamples++
			}
		}
	}

	if totalSamples == 0 {
		return 0.0
	}

	return math.Min(highFreqNoise/float64(totalSamples), 1.0)
}

// isSmoothArea checks if an area is relatively smooth (low gradient)
func (c *CompressionAnalyzer) isSmoothArea(gray *image.Gray, x, y, radius int) bool {
	r, _, _, _ := gray.At(x, y).RGBA()
	center := r
	maxGradient := 0.0

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			nr, _, _, _ := gray.At(x+dx, y+dy).RGBA()
			neighbor := nr

			gradient := math.Abs(float64(center) - float64(neighbor))
			if gradient > maxGradient {
				maxGradient = gradient
			}
		}
	}

	return maxGradient < 5000 // Threshold for smooth area
}

// measureHighFrequencyNoise measures high-frequency variations
func (c *CompressionAnalyzer) measureHighFrequencyNoise(gray *image.Gray, x, y int) float64 {
	var noise float64
	samples := 0

	// Read center pixel (RGBA returns 4 values, only R is needed here)
	r, _, _, _ := gray.At(x, y).RGBA()
	center := r

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			nr, _, _, _ := gray.At(x+dx, y+dy).RGBA()
			neighbor := nr

			variation := math.Abs(float64(center) - float64(neighbor))
			if variation > 2000 { // High-frequency noise threshold
				noise += 1.0
			}
			samples++
		}
	}

	if samples == 0 {
		return 0.0
	}

	return noise / float64(samples)
}
