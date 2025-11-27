package quality

import (
	"image"
	"math"

	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/sirupsen/logrus"
)

// Analyzer performs comprehensive image quality assessment
type Analyzer struct {
	config Config
	logger *logrus.Logger
}

// Config defines quality analysis parameters and thresholds
type Config struct {
	DetailedAnalysis   bool
	SharpnessThreshold float64
	NoiseThreshold     float64
	MinExposure        float64
	MaxExposure        float64
	MinContrast        float64
	CompressionQuality float64
}

// DefaultConfig returns sensible default quality analysis configuration
func DefaultConfig() Config {
	return Config{
		DetailedAnalysis:   true,
		SharpnessThreshold: 0.1,
		NoiseThreshold:     0.3,
		MinExposure:        0.1,
		MaxExposure:        0.9,
		MinContrast:        0.2,
		CompressionQuality: 0.8,
	}
}

// NewAnalyzer creates a new image quality analyzer
func NewAnalyzer(cfg Config) *Analyzer {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Analyzer{
		config: cfg,
		logger: logger,
	}
}

// Analyze performs comprehensive quality assessment on an image
func (a *Analyzer) Analyze(img image.Image) (*api.ImageQuality, error) {
	quality := &api.ImageQuality{}

	// Convert to grayscale for some analyses
	gray := toGray(img)

	// Perform individual quality analyses
	var err error
	quality.Sharpness, err = a.analyzeSharpness(gray)
	if err != nil {
		a.logger.Warnf("Sharpness analysis failed: %v", err)
	}

	quality.Noise, err = a.analyzeNoise(gray)
	if err != nil {
		a.logger.Warnf("Noise analysis failed: %v", err)
	}

	quality.Exposure, err = a.analyzeExposure(gray)
	if err != nil {
		a.logger.Warnf("Exposure analysis failed: %v", err)
	}

	quality.Contrast, err = a.analyzeContrast(gray)
	if err != nil {
		a.logger.Warnf("Contrast analysis failed: %v", err)
	}

	quality.Compression, err = a.analyzeCompression(img)
	if err != nil {
		a.logger.Warnf("Compression analysis failed: %v", err)
	}

	quality.ColorCast, err = a.analyzeColorCast(img)
	if err != nil {
		a.logger.Warnf("Color cast analysis failed: %v", err)
	}

	// Calculate final composite score
	quality.FinalScore = a.calculateFinalScore(quality)

	a.logger.Debugf("Quality analysis completed: Sharpness=%.2f, Noise=%.2f, Final=%.1f",
		quality.Sharpness, quality.Noise, quality.FinalScore)

	return quality, nil
}

func toGray(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	return gray
}

// analyzeSharpness calculates image sharpness using Laplacian variance method
func (a *Analyzer) analyzeSharpness(gray *image.Gray) (float64, error) {
	bounds := gray.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	if width < 3 || height < 3 {
		return 0, api.ErrImageTooSmall
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
	normalized := math.Min(variance/100.0, 1.0) // Assuming max variance around 100

	return normalized, nil
}

// analyzeNoise estimates image noise level
func (a *Analyzer) analyzeNoise(gray *image.Gray) (float64, error) {
	bounds := gray.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	if width < 3 || height < 3 {
		return 0, api.ErrImageTooSmall
	}

	var noiseSum float64
	var count int

	// Analyze local variance in smooth regions
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			center := float64(gray.GrayAt(x, y).Y)
			neighbors := []float64{
				float64(gray.GrayAt(x-1, y-1).Y),
				float64(gray.GrayAt(x, y-1).Y),
				float64(gray.GrayAt(x+1, y-1).Y),
				float64(gray.GrayAt(x-1, y).Y),
				float64(gray.GrayAt(x+1, y).Y),
				float64(gray.GrayAt(x-1, y+1).Y),
				float64(gray.GrayAt(x, y+1).Y),
				float64(gray.GrayAt(x+1, y+1).Y),
			}

			// Calculate local variance
			var mean, variance float64
			for _, neighbor := range neighbors {
				mean += neighbor
			}
			mean /= float64(len(neighbors))

			for _, neighbor := range neighbors {
				diff := neighbor - mean
				variance += diff * diff
			}
			variance /= float64(len(neighbors))

			// Only consider smooth regions (low gradient)
			gradient := math.Abs(center - mean)
			if gradient < 10 { // Threshold for smooth regions
				noiseSum += math.Sqrt(variance)
				count++
			}
		}
	}

	if count == 0 {
		return 0.5, nil // Default medium noise if no smooth regions found
	}

	avgNoise := noiseSum / float64(count)
	normalized := math.Min(avgNoise/50.0, 1.0) // Normalize to 0-1

	return normalized, nil
}

// analyzeExposure assesses image exposure level
func (a *Analyzer) analyzeExposure(gray *image.Gray) (float64, error) {
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
			luminance := float64(gray.GrayAt(x, y).Y) / 255.0
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

// analyzeContrast measures image contrast using standard deviation
func (a *Analyzer) analyzeContrast(gray *image.Gray) (float64, error) {
	bounds := gray.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()

	if totalPixels == 0 {
		return 0.5, nil
	}

	var sum, sumSquares float64

	// Calculate mean and standard deviation
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			value := float64(gray.GrayAt(x, y).Y) / 255.0
			sum += value
			sumSquares += value * value
		}
	}

	mean := sum / float64(totalPixels)
	variance := (sumSquares / float64(totalPixels)) - (mean * mean)
	stdDev := math.Sqrt(math.Max(0, variance))

	// Normalize contrast score (good contrast around 0.2-0.3 std dev)
	contrast := math.Min(stdDev/0.4, 1.0)

	return contrast, nil
}

// analyzeCompression detects compression artifacts (simplified)
func (a *Analyzer) analyzeCompression(img image.Image) (float64, error) {
	// Simplified compression artifact detection
	// In a real implementation, this would analyze blocking artifacts
	// common in JPEG compression

	bounds := img.Bounds()
	if bounds.Dx() < 100 || bounds.Dy() < 100 {
		return 0.1, nil // Small images have less noticeable compression
	}

	// Placeholder: assume moderate compression for most images
	// Real implementation would analyze high-frequency noise patterns
	return 0.3, nil
}

// analyzeColorCast detects color balance issues
func (a *Analyzer) analyzeColorCast(img image.Image) (float64, error) {
	bounds := img.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()

	if totalPixels == 0 {
		return 0, nil
	}

	var rSum, gSum, bSum float64

	// Calculate average RGB values
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 4 { // Sample every 4th pixel for performance
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

	return deviation, nil
}

// calculateFinalScore computes a composite quality score from individual metrics
func (a *Analyzer) calculateFinalScore(quality *api.ImageQuality) float64 {
	// Weighted combination of quality factors
	weights := map[string]float64{
		"sharpness":   0.3,  // Sharpness is most important
		"noise":       0.25, // Noise reduction important
		"exposure":    0.2,  // Proper exposure
		"contrast":    0.15, // Good contrast
		"compression": 0.05, // Compression artifacts
		"color_cast":  0.05, // Color balance
	}

	// Adjust exposure score to penalize both under and overexposure
	exposureScore := 1.0 - math.Abs(quality.Exposure-0.5)*2

	// Calculate weighted score
	score := quality.Sharpness*weights["sharpness"] +
		(1-quality.Noise)*weights["noise"] + // Invert noise (lower is better)
		exposureScore*weights["exposure"] +
		quality.Contrast*weights["contrast"] +
		(1-quality.Compression)*weights["compression"] + // Invert compression
		(1-quality.ColorCast)*weights["color_cast"] // Invert color cast

	// Convert to 0-100 scale
	finalScore := score * 100

	// Ensure valid range
	return math.Max(0, math.Min(100, finalScore))
}

// IsBlurry determines if an image is blurry based on sharpness threshold
func (a *Analyzer) IsBlurry(quality api.ImageQuality) bool {
	return quality.Sharpness < a.config.SharpnessThreshold
}

// IsOverexposed determines if an image is overexposed
func (a *Analyzer) IsOverexposed(quality api.ImageQuality) bool {
	return quality.Exposure > a.config.MaxExposure
}

// IsUnderexposed determines if an image is underexposed
func (a *Analyzer) IsUnderexposed(quality api.ImageQuality) bool {
	return quality.Exposure < a.config.MinExposure
}

// IsLowQuality determines if an image has overall low quality
func (a *Analyzer) IsLowQuality(quality api.ImageQuality) bool {
	return quality.FinalScore < 50.0
}
