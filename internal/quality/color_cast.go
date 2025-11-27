package quality

import (
	"fmt"
	"image"
	"math"
)

// ColorCastAnalyzer analyzes color balance and cast in images
type ColorCastAnalyzer struct{}

// NewColorCastAnalyzer creates a new color cast analyzer
func NewColorCastAnalyzer() *ColorCastAnalyzer {
	return &ColorCastAnalyzer{}
}

// AnalyzeColorCast detects color balance issues in an image
func (c *ColorCastAnalyzer) AnalyzeColorCast(img image.Image) (float64, error) {
	bounds := img.Bounds()
	totalPixels := float64(bounds.Dx() * bounds.Dy())

	if totalPixels == 0 {
		return 0, nil
	}

	var rSum, gSum, bSum float64

	// Calculate average RGB values
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 4 { // Sample for performance
		for x := bounds.Min.X; x < bounds.Max.X; x += 4 {
			r, g, b, _ := img.At(x, y).RGBA()
			rSum += float64(r) / 65535.0
			gSum += float64(g) / 65535.0
			bSum += float64(b) / 65535.0
		}
	}

	sampledPixels := float64(((bounds.Dx() / 4) + 1) * ((bounds.Dy() / 4) + 1))
	rAvg := rSum / sampledPixels
	gAvg := gSum / sampledPixels
	bAvg := bSum / sampledPixels

	// Calculate color cast as deviation from neutral gray
	maxChannel := math.Max(math.Max(rAvg, gAvg), bAvg)
	if maxChannel == 0 {
		return 0, nil
	}

	rNorm := rAvg / maxChannel
	gNorm := gAvg / maxChannel
	bNorm := bAvg / maxChannel

	// Measure deviation from balanced white (1,1,1)
	deviation := math.Sqrt(
		math.Pow(1.0-rNorm, 2)+
			math.Pow(1.0-gNorm, 2)+
			math.Pow(1.0-bNorm, 2)) / math.Sqrt(3)

	return math.Min(deviation, 1.0), nil
}

// DetectColorTemperature estimates the color temperature of an image
func (c *ColorCastAnalyzer) DetectColorTemperature(img image.Image) (string, float64) {
	bounds := img.Bounds()
	var rSum, bSum float64
	var count int

	// Calculate red-blue ratio for temperature estimation
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 8 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 8 {
			r, _, b, _ := img.At(x, y).RGBA()
			rSum += float64(r)
			bSum += float64(b)
			count++
		}
	}

	if count == 0 || bSum == 0 {
		return "neutral", 0.5
	}

	rbRatio := rSum / bSum

	var temperature string
	var score float64

	if rbRatio > 1.2 {
		temperature = "warm"
		score = math.Min((rbRatio-1.2)/0.8, 1.0)
	} else if rbRatio < 0.8 {
		temperature = "cool"
		score = math.Min((0.8-rbRatio)/0.8, 1.0)
	} else {
		temperature = "neutral"
		score = 0.0
	}

	return temperature, score
}

// AnalyzeColorDistribution analyzes the overall color distribution
func (c *ColorCastAnalyzer) AnalyzeColorDistribution(img image.Image) map[string]float64 {
	bounds := img.Bounds()
	colorStats := make(map[string]float64)

	var saturatedPixels, mutedPixels, darkPixels, brightPixels int
	totalPixels := bounds.Dx() * bounds.Dy()
	fmt.Print(totalPixels)
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 4 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 4 {
			r, g, b, _ := img.At(x, y).RGBA()

			rNorm := float64(r) / 65535.0
			gNorm := float64(g) / 65535.0
			bNorm := float64(b) / 65535.0

			// Check saturation
			maxRGB := math.Max(math.Max(rNorm, gNorm), bNorm)
			minRGB := math.Min(math.Min(rNorm, gNorm), bNorm)
			saturation := 0.0
			if maxRGB > 0 {
				saturation = (maxRGB - minRGB) / maxRGB
			}

			// Check brightness
			brightness := (rNorm + gNorm + bNorm) / 3

			if saturation > 0.7 {
				saturatedPixels++
			} else if saturation < 0.3 {
				mutedPixels++
			}

			if brightness < 0.2 {
				darkPixels++
			} else if brightness > 0.8 {
				brightPixels++
			}
		}
	}

	sampledPixels := ((bounds.Dx() / 4) + 1) * ((bounds.Dy() / 4) + 1)
	colorStats["saturation_high"] = float64(saturatedPixels) / float64(sampledPixels)
	colorStats["saturation_low"] = float64(mutedPixels) / float64(sampledPixels)
	colorStats["brightness_low"] = float64(darkPixels) / float64(sampledPixels)
	colorStats["brightness_high"] = float64(brightPixels) / float64(sampledPixels)

	return colorStats
}
