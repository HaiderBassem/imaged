package imaging

import (
	"image"
	"math"

	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
)

// Preprocessor handles image preprocessing for consistent analysis
type Preprocessor struct {
	maxDimension int
	quality      int
}

// NewPreprocessor creates a new image preprocessor
func NewPreprocessor(maxDimension, quality int) *Preprocessor {
	return &Preprocessor{
		maxDimension: maxDimension,
		quality:      quality,
	}
}

// PreprocessImage prepares an image for analysis
func (p *Preprocessor) PreprocessImage(img image.Image) image.Image {
	// Step 1: Resize if too large
	img = p.resizeImage(img)

	// Step 2: Normalize orientation
	img = p.normalizeOrientation(img)

	// Step 3: Enhance for analysis (optional)
	img = p.enhanceImage(img)

	return img
}

// resizeImage resizes image while maintaining aspect ratio
func (p *Preprocessor) resizeImage(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Check if resizing is needed
	if width <= p.maxDimension && height <= p.maxDimension {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	var newWidth, newHeight int
	if width > height {
		newWidth = p.maxDimension
		newHeight = int(float64(height) * float64(p.maxDimension) / float64(width))
	} else {
		newHeight = p.maxDimension
		newWidth = int(float64(width) * float64(p.maxDimension) / float64(height))
	}

	// Use high-quality resizing
	return resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)
}

// normalizeOrientation handles image orientation
func (p *Preprocessor) normalizeOrientation(img image.Image) image.Image {
	// This is a simplified implementation
	// In production, you would read EXIF orientation and rotate accordingly
	return img
}

// enhanceImage applies enhancements for better analysis
func (p *Preprocessor) enhanceImage(img image.Image) image.Image {
	// Convert to grayscale for some analyses
	// Note: Return color image for color-based analyses
	return imaging.Grayscale(img)
}

// ExtractROI extracts region of interest from image
func (p *Preprocessor) ExtractROI(img image.Image, x, y, width, height int) image.Image {
	bounds := img.Bounds()

	// Ensure ROI is within bounds
	x = max(bounds.Min.X, x)
	y = max(bounds.Min.Y, y)
	width = min(width, bounds.Max.X-x)
	height = min(height, bounds.Max.Y-y)

	// Extract sub-image
	roi := imaging.Crop(img, image.Rect(x, y, x+width, y+height))
	return roi
}

// ComputeImageMoments calculates image moments for shape analysis
func (p *Preprocessor) ComputeImageMoments(img image.Image) map[string]float64 {
	bounds := img.Bounds()
	gray := imaging.Grayscale(img)

	var m00, m10, m01, m11, m20, m02, m30, m03, m12, m21 float64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, _, _, _ := gray.At(x, y).RGBA()
			intensity := float64(r) / 65535.0

			xf := float64(x)
			yf := float64(y)

			// Spatial moments
			m00 += intensity
			m10 += xf * intensity
			m01 += yf * intensity
			m11 += xf * yf * intensity
			m20 += xf * xf * intensity
			m02 += yf * yf * intensity
			m30 += xf * xf * xf * intensity
			m03 += yf * yf * yf * intensity
			m12 += xf * yf * yf * intensity
			m21 += xf * xf * yf * intensity
		}
	}

	if m00 == 0 {
		return map[string]float64{
			"hu1": 0, "hu2": 0, "hu3": 0,
			"centroid_x": 0,
			"centroid_y": 0,
		}
	}

	// Centroid
	cx := m10 / m00
	cy := m01 / m00

	// Central moments
	mu11 := m11 - (m10*m01)/m00
	mu20 := m20 - (m10*m10)/m00
	mu02 := m02 - (m01*m01)/m00
	mu30 := m30 - (3 * m20 * cx) + (2 * m10 * cx * cx)
	mu03 := m03 - (3 * m02 * cy) + (2 * m01 * cy * cy)
	mu12 := m12 - (2 * m11 * cy) - (m01*m20)/m00 + (2 * m10 * cy * cy)
	mu21 := m21 - (2 * m11 * cx) - (m10*m02)/m00 + (2 * m01 * cx * cx)

	// Hu invariant moments
	hu1 := mu20 + mu02
	hu2 := math.Pow(mu20-mu02, 2) + 4*math.Pow(mu11, 2)
	hu3 := math.Pow(mu30-3*mu12, 2) + math.Pow(3*mu21-mu03, 2)

	return map[string]float64{
		"hu1":        hu1,
		"hu2":        hu2,
		"hu3":        hu3,
		"centroid_x": cx,
		"centroid_y": cy,
	}
}

// CreateImagePyramid creates multi-scale image pyramid
func (p *Preprocessor) CreateImagePyramid(img image.Image, levels int) []image.Image {
	pyramid := make([]image.Image, levels)
	pyramid[0] = img

	for i := 1; i < levels; i++ {
		// Reduce size by half each level
		bounds := pyramid[i-1].Bounds()
		newWidth := bounds.Dx() / 2
		newHeight := bounds.Dy() / 2

		if newWidth < 16 || newHeight < 16 {
			break // Stop if image becomes too small
		}

		pyramid[i] = resize.Resize(uint(newWidth), uint(newHeight), pyramid[i-1], resize.Lanczos3)
	}

	return pyramid[:levels]
}

// Utility functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
