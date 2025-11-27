package utils

import (
	"image"

	"github.com/disintegration/imaging"
)

// ImageUtils provides common image manipulation utilities
type ImageUtils struct{}

// NewImageUtils creates a new image utilities instance
func NewImageUtils() *ImageUtils {
	return &ImageUtils{}
}

// CalculateAspectRatio calculates the aspect ratio of an image
func (iu *ImageUtils) CalculateAspectRatio(img image.Image) float64 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if height == 0 {
		return 0
	}
	return float64(width) / float64(height)
}

// IsPortrait checks if an image is in portrait orientation
func (iu *ImageUtils) IsPortrait(img image.Image) bool {
	bounds := img.Bounds()
	return bounds.Dy() > bounds.Dx()
}

// IsLandscape checks if an image is in landscape orientation
func (iu *ImageUtils) IsLandscape(img image.Image) bool {
	bounds := img.Bounds()
	return bounds.Dx() > bounds.Dy()
}

// CalculateFileSize returns the estimated file size in bytes
func (iu *ImageUtils) CalculateFileSize(img image.Image, format string) int64 {
	bounds := img.Bounds()
	pixels := bounds.Dx() * bounds.Dy()

	// Rough estimation based on format and image size
	switch format {
	case "jpeg", "jpg":
		return int64(float64(pixels) * 0.3) // ~0.3 bytes per pixel for JPEG
	case "png":
		return int64(float64(pixels) * 1.5) // ~1.5 bytes per pixel for PNG
	case "webp":
		return int64(float64(pixels) * 0.4) // ~0.4 bytes per pixel for WebP
	default:
		return int64(float64(pixels) * 1.0) // Default estimation
	}
}

// ComputeBrightnessHistogram calculates brightness distribution
func (iu *ImageUtils) ComputeBrightnessHistogram(img image.Image, bins int) []float64 {
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

	return histogram
}

// DetectBlankImage checks if an image is mostly blank/empty
func (iu *ImageUtils) DetectBlankImage(img image.Image, threshold float64) bool {
	bounds := img.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()

	if totalPixels == 0 {
		return true
	}

	gray := imaging.Grayscale(img)
	var uniformPixels int

	// Sample pixels to check for uniformity
	sampleRate := 10
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleRate {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleRate {
			r, _, _, _ := gray.At(x, y).RGBA()
			// Consider pixel as "blank" if it's near white or black
			if r < 0x1000 || r > 0xF000 {
				uniformPixels++
			}
		}
	}

	sampledPixels := (bounds.Dx() / sampleRate) * (bounds.Dy() / sampleRate)
	blankRatio := float64(uniformPixels) / float64(sampledPixels)

	return blankRatio >= threshold
}
