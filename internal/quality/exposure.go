package quality

import (
	"image"
	"math"

	"github.com/disintegration/imaging"
)

// ExposureAnalyzer analyzes image exposure levels
type ExposureAnalyzer struct{}

// NewExposureAnalyzer creates a new exposure analyzer
func NewExposureAnalyzer() *ExposureAnalyzer {
	return &ExposureAnalyzer{}
}

// AnalyzeExposure assesses image exposure level
func (e *ExposureAnalyzer) AnalyzeExposure(img image.Image) (float64, error) {
	gray := imaging.Grayscale(img)
	bounds := gray.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()

	if totalPixels == 0 {
		return 0.5, nil // Default neutral exposure
	}

	var sum float64
	var darkPixels, brightPixels int

	// Analyze histogram for exposure characteristics
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, _, _, _ := gray.At(x, y).RGBA()
			luminance := float64(r) / 65535.0
			sum += luminance

			if luminance < 0.1 {
				darkPixels++
			} else if luminance > 0.9 {
				brightPixels++
			}
		}
	}

	avgLuminance := sum / float64(totalPixels)

	// Penalize images with too many dark or bright pixels
	darkRatio := float64(darkPixels) / float64(totalPixels)
	brightRatio := float64(brightPixels) / float64(totalPixels)

	// Adjust exposure score based on distribution
	exposure := avgLuminance
	if darkRatio > 0.3 {
		exposure -= (darkRatio - 0.3) * 0.5
	}
	if brightRatio > 0.3 {
		exposure += (brightRatio - 0.3) * 0.5
	}

	// Clamp to valid range
	exposure = math.Max(0, math.Min(1, exposure))

	return exposure, nil
}

// IsOverexposed checks if image is overexposed
func (e *ExposureAnalyzer) IsOverexposed(img image.Image, threshold float64) (bool, error) {
	exposure, err := e.AnalyzeExposure(img)
	if err != nil {
		return false, err
	}
	return exposure > threshold, nil
}

// IsUnderexposed checks if image is underexposed
func (e *ExposureAnalyzer) IsUnderexposed(img image.Image, threshold float64) (bool, error) {
	exposure, err := e.AnalyzeExposure(img)
	if err != nil {
		return false, err
	}
	return exposure < threshold, nil
}

// GetExposureHistogram returns the luminance histogram
func (e *ExposureAnalyzer) GetExposureHistogram(img image.Image, bins int) ([]float64, error) {
	gray := imaging.Grayscale(img)
	bounds := gray.Bounds()
	histogram := make([]float64, bins)
	totalPixels := float64(bounds.Dx() * bounds.Dy())

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, _, _, _ := gray.At(x, y).RGBA()
			brightness := float64(r) / 65535.0
			bin := int(brightness * float64(bins-1))
			if bin >= 0 && bin < bins {
				histogram[bin]++
			}
		}
	}

	// Normalize
	for i := range histogram {
		histogram[i] /= totalPixels
	}

	return histogram, nil
}
