package hash

import (
	"image"
	"math"
)

// ColorSignature computes color-based signatures for images
type ColorSignature struct {
	Bins int
}

// NewColorSignature creates a new color signature calculator
func NewColorSignature(bins int) *ColorSignature {
	return &ColorSignature{
		Bins: bins,
	}
}

// ComputeColorHistogram calculates RGB color histogram
func (c *ColorSignature) ComputeColorHistogram(img image.Image) ([]float64, error) {
	bounds := img.Bounds()
	totalPixels := float64(bounds.Dx() * bounds.Dy())
	histogram := make([]float64, c.Bins*3) // R, G, B channels

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			// Convert to 0-1 range and bin
			rNorm := float64(r) / 65535.0
			gNorm := float64(g) / 65535.0
			bNorm := float64(b) / 65535.0

			rBin := int(rNorm * float64(c.Bins-1))
			gBin := int(gNorm * float64(c.Bins-1))
			bBin := int(bNorm * float64(c.Bins-1))

			if rBin >= 0 && rBin < c.Bins {
				histogram[rBin]++
			}
			if gBin >= 0 && gBin < c.Bins {
				histogram[c.Bins+gBin]++
			}
			if bBin >= 0 && bBin < c.Bins {
				histogram[2*c.Bins+bBin]++
			}
		}
	}

	// Normalize
	for i := range histogram {
		histogram[i] /= totalPixels
	}

	return histogram, nil
}

// ComputeHSVHistogram calculates HSV color histogram
func (c *ColorSignature) ComputeHSVHistogram(img image.Image) ([]float64, error) {
	bounds := img.Bounds()
	totalPixels := float64(bounds.Dx() * bounds.Dy())
	histogram := make([]float64, c.Bins*3) // H, S, V channels

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			// Convert to HSV
			h, s, v := rgbToHSV(
				float64(r)/65535.0,
				float64(g)/65535.0,
				float64(b)/65535.0,
			)

			hBin := int(h * float64(c.Bins-1))
			sBin := int(s * float64(c.Bins-1))
			vBin := int(v * float64(c.Bins-1))

			if hBin >= 0 && hBin < c.Bins {
				histogram[hBin]++
			}
			if sBin >= 0 && sBin < c.Bins {
				histogram[c.Bins+sBin]++
			}
			if vBin >= 0 && vBin < c.Bins {
				histogram[2*c.Bins+vBin]++
			}
		}
	}

	// Normalize
	for i := range histogram {
		histogram[i] /= totalPixels
	}

	return histogram, nil
}

// CompareHistograms compares two histograms using Earth Mover's Distance
func (c *ColorSignature) CompareHistograms(hist1, hist2 []float64) float64 {
	if len(hist1) != len(hist2) {
		return 1.0
	}

	// Use Chi-squared distance for color histograms
	var sum float64
	for i := range hist1 {
		if hist1[i]+hist2[i] > 0 {
			diff := hist1[i] - hist2[i]
			sum += (diff * diff) / (hist1[i] + hist2[i])
		}
	}

	distance := sum / 2.0
	// Convert to similarity (0-1)
	similarity := 1.0 / (1.0 + distance)
	return math.Max(0.0, math.Min(1.0, similarity))
}

// rgbToHSV converts RGB to HSV color space
func rgbToHSV(r, g, b float64) (h, s, v float64) {
	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)

	v = max

	if max == 0 {
		s = 0
	} else {
		s = (max - min) / max
	}

	if max == min {
		h = 0
	} else {
		delta := max - min
		switch max {
		case r:
			h = (g - b) / delta
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/delta + 2
		case b:
			h = (r-g)/delta + 4
		}
		h /= 6
	}

	return h, s, v
}
