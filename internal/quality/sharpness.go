package quality

import (
	"image"
	"math"

	"github.com/disintegration/imaging"
)

// SharpnessAnalyzer measures image sharpness and blur
type SharpnessAnalyzer struct{}

// NewSharpnessAnalyzer creates a new sharpness analyzer
func NewSharpnessAnalyzer() *SharpnessAnalyzer {
	return &SharpnessAnalyzer{}
}

// AnalyzeSharpness calculates image sharpness using Laplacian variance
func (s *SharpnessAnalyzer) AnalyzeSharpness(img image.Image) (float64, error) {
	_ = imaging.Grayscale(img) // Keep this to preserve your original structure/intent

	// Convert to true grayscale image (required for GrayAt)
	gray := image.NewGray(img.Bounds())
	for y := gray.Bounds().Min.Y; y < gray.Bounds().Max.Y; y++ {
		for x := gray.Bounds().Min.X; x < gray.Bounds().Max.X; x++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	bounds := gray.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	if width < 3 || height < 3 {
		return 0, nil
	}

	var sum float64
	var count int

	// Apply simple Laplacian kernel to detect edges
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			// Simple Laplacian: center * 4 - neighbors
			center := float64(gray.GrayAt(x, y).Y)
			top := float64(gray.GrayAt(x, y-1).Y)
			bottom := float64(gray.GrayAt(x, y+1).Y)
			left := float64(gray.GrayAt(x-1, y).Y)
			right := float64(gray.GrayAt(x+1, y).Y)

			laplacian := math.Abs(4*center - (top + bottom + left + right))
			sum += laplacian
			count++
		}
	}

	if count == 0 {
		return 0, nil
	}

	// Normalize variance to 0-1 range
	variance := sum / float64(count)

	// NOTE:
	// The divisor (100.0) is an empirical value that may be tuned
	// depending on typical image contrast and resolution.
	normalized := math.Min(variance/100.0, 1.0)

	return normalized, nil
}

// IsBlurry determines if an image is blurry based on sharpness threshold
func (s *SharpnessAnalyzer) IsBlurry(img image.Image, threshold float64) (bool, error) {
	sharpness, err := s.AnalyzeSharpness(img)
	if err != nil {
		return false, err
	}
	return sharpness < threshold, nil
}

// AnalyzeEdgeStrength analyzes edge strength distribution
func (s *SharpnessAnalyzer) AnalyzeEdgeStrength(img image.Image) (map[string]float64, error) {
	_ = imaging.Grayscale(img) // Preserve original intent

	// Convert to true grayscale image
	gray := image.NewGray(img.Bounds())
	for y := gray.Bounds().Min.Y; y < gray.Bounds().Max.Y; y++ {
		for x := gray.Bounds().Min.X; x < gray.Bounds().Max.X; x++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	bounds := gray.Bounds()

	edgeStats := make(map[string]float64)
	var strongEdges, mediumEdges, weakEdges int
	totalPixels := bounds.Dx() * bounds.Dy()

	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			// Calculate gradient magnitude using Sobel operator
			gx, gy := s.sobelGradient(gray, x, y)
			magnitude := math.Sqrt(gx*gx + gy*gy)

			// Classify edge strength
			if magnitude > 50 {
				strongEdges++
			} else if magnitude > 20 {
				mediumEdges++
			} else if magnitude > 5 {
				weakEdges++
			}
		}
	}

	// Normalize counts to image area
	edgeStats["strong_edges"] = float64(strongEdges) / float64(totalPixels)
	edgeStats["medium_edges"] = float64(mediumEdges) / float64(totalPixels)
	edgeStats["weak_edges"] = float64(weakEdges) / float64(totalPixels)
	edgeStats["total_edges"] = float64(strongEdges+mediumEdges+weakEdges) / float64(totalPixels)

	return edgeStats, nil
}

// sobelGradient calculates Sobel gradient for a pixel
func (s *SharpnessAnalyzer) sobelGradient(gray *image.Gray, x, y int) (float64, float64) {
	// Sobel kernels
	kernelX := [3][3]float64{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}
	kernelY := [3][3]float64{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}

	var gx, gy float64

	for ky := -1; ky <= 1; ky++ {
		for kx := -1; kx <= 1; kx++ {
			pixel := float64(gray.GrayAt(x+kx, y+ky).Y)
			gx += pixel * kernelX[ky+1][kx+1]
			gy += pixel * kernelY[ky+1][kx+1]
		}
	}

	return gx, gy
}

// AnalyzeFocusRegion analyzes sharpness in different image regions
func (s *SharpnessAnalyzer) AnalyzeFocusRegion(img image.Image) (map[string]float64, error) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	regions := make(map[string]float64)

	// Define regions: center, corners, edges
	regions["center"] = s.analyzeRegionSharpness(img, width/4, height/4, width/2, height/2)
	regions["top_left"] = s.analyzeRegionSharpness(img, 0, 0, width/4, height/4)
	regions["top_right"] = s.analyzeRegionSharpness(img, 3*width/4, 0, width/4, height/4)
	regions["bottom_left"] = s.analyzeRegionSharpness(img, 0, 3*height/4, width/4, height/4)
	regions["bottom_right"] = s.analyzeRegionSharpness(img, 3*width/4, 3*height/4, width/4, height/4)

	return regions, nil
}

// analyzeRegionSharpness analyzes sharpness in a specific region
func (s *SharpnessAnalyzer) analyzeRegionSharpness(img image.Image, x, y, width, height int) float64 {
	// Extract region safely
	region := imaging.Crop(img, image.Rect(x, y, x+width, y+height))

	sharpness, err := s.AnalyzeSharpness(region)
	if err != nil {
		// In case of failure, return neutral sharpness score
		return 0.0
	}
	return sharpness
}
