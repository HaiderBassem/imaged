package imaging

import (
	"image"
	"image/color"
	"math"
)

// ColorSpace provides color space conversion utilities
type ColorSpace struct{}

// NewColorSpace creates a new color space converter
func NewColorSpace() *ColorSpace {
	return &ColorSpace{}
}

// RGBToHSV converts RGB color to HSV color space
func (cs *ColorSpace) RGBToHSV(r, g, b uint32) (h, s, v float64) {
	rf := float64(r) / 65535.0
	gf := float64(g) / 65535.0
	bf := float64(b) / 65535.0

	max := math.Max(math.Max(rf, gf), bf)
	min := math.Min(math.Min(rf, gf), bf)

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
		case rf:
			h = (gf - bf) / delta
			if gf < bf {
				h += 6
			}
		case gf:
			h = (bf-rf)/delta + 2
		case bf:
			h = (rf-gf)/delta + 4
		}
		h /= 6
	}

	return h, s, v
}

// HSVToRGB converts HSV color to RGB color space
func (cs *ColorSpace) HSVToRGB(h, s, v float64) (r, g, b uint32) {
	if s == 0 {
		// Achromatic (gray)
		gray := uint32(v * 65535.0)
		return gray, gray, gray
	}

	h = h * 6 // sector 0 to 5
	i := math.Floor(h)
	f := h - i // fractional part of h
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	var rf, gf, bf float64
	switch int(i) % 6 {
	case 0:
		rf, gf, bf = v, t, p
	case 1:
		rf, gf, bf = q, v, p
	case 2:
		rf, gf, bf = p, v, t
	case 3:
		rf, gf, bf = p, q, v
	case 4:
		rf, gf, bf = t, p, v
	case 5:
		rf, gf, bf = v, p, q
	}

	r = uint32(rf * 65535.0)
	g = uint32(gf * 65535.0)
	b = uint32(bf * 65535.0)

	return r, g, b
}

// RGBToLAB converts RGB color to LAB color space
func (cs *ColorSpace) RGBToLAB(r, g, b uint32) (l, a, b2 float64) {
	// Convert to XYZ first
	x, y, z := cs.RGBToXYZ(r, g, b)

	// Convert to LAB
	return cs.XYZToLAB(x, y, z)
}

// RGBToXYZ converts RGB color to XYZ color space
func (cs *ColorSpace) RGBToXYZ(r, g, b uint32) (x, y, z float64) {
	// Normalize to 0-1
	rf := float64(r) / 65535.0
	gf := float64(g) / 65535.0
	bf := float64(b) / 65535.0

	// Apply gamma correction
	rf = cs.gammaCorrection(rf)
	gf = cs.gammaCorrection(gf)
	bf = cs.gammaCorrection(bf)

	// Convert to XYZ using sRGB matrix
	x = 0.4124564*rf + 0.3575761*gf + 0.1804375*bf
	y = 0.2126729*rf + 0.7151522*gf + 0.0721750*bf
	z = 0.0193339*rf + 0.1191920*gf + 0.9503041*bf

	return x * 100, y * 100, z * 100
}

// XYZToLAB converts XYZ color to LAB color space
func (cs *ColorSpace) XYZToLAB(x, y, z float64) (l, a, b2 float64) {
	// Reference white (D65)
	refX, refY, refZ := 95.047, 100.000, 108.883

	// Normalize by reference white
	x = x / refX
	y = y / refY
	z = z / refZ

	// Apply nonlinear transformation
	x = cs.nonlinearTransform(x)
	y = cs.nonlinearTransform(y)
	z = cs.nonlinearTransform(z)

	// Calculate LAB components
	l = 116*y - 16
	a = 500 * (x - y)
	b2 = 200 * (y - z)

	return l, a, b2
}

// gammaCorrection applies sRGB gamma correction
func (cs *ColorSpace) gammaCorrection(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// nonlinearTransform applies LAB nonlinear transformation
func (cs *ColorSpace) nonlinearTransform(t float64) float64 {
	if t > 0.008856 {
		return math.Pow(t, 1.0/3.0)
	}
	return 7.787*t + 16.0/116.0
}

// ColorDistance calculates the distance between two colors in LAB space
func (cs *ColorSpace) ColorDistance(c1, c2 color.Color) float64 {
	r1, g1, b1u, _ := c1.RGBA()
	r2, g2, b2u, _ := c2.RGBA()

	l1, a1, bb1 := cs.RGBToLAB(r1, g1, b1u)
	l2, a2, bb2 := cs.RGBToLAB(r2, g2, b2u)

	// Calculate Delta E (CIE76)
	dl := l1 - l2
	da := a1 - a2
	db := bb1 - bb2

	return math.Sqrt(dl*dl + da*da + db*db)
}

// IsSimilarColor checks if two colors are similar within a threshold
func (cs *ColorSpace) IsSimilarColor(c1, c2 color.Color, threshold float64) bool {
	distance := cs.ColorDistance(c1, c2)
	return distance <= threshold
}

// GetDominantColors extracts dominant colors from an image (simplified)
func (cs *ColorSpace) GetDominantColors(img image.Image, maxColors int) []color.Color {
	// This is a simplified implementation
	// In production, use proper color quantization algorithms

	bounds := img.Bounds()
	colorCount := make(map[color.RGBA]int)

	// Sample pixels for color analysis
	sampleRate := 10
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleRate {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleRate {
			c := img.At(x, y)
			rgba := color.RGBAModel.Convert(c).(color.RGBA)

			// Quantize color to reduce variations
			quantized := cs.quantizeColor(rgba, 16)
			colorCount[quantized]++
		}
	}

	// Find most frequent colors
	var colors []color.Color
	for c, count := range colorCount {
		if count > len(colorCount)/20 { // At least 5% frequency
			colors = append(colors, c)
		}
	}

	// Limit to maxColors
	if len(colors) > maxColors {
		colors = colors[:maxColors]
	}

	return colors
}

// quantizeColor reduces color precision
func (cs *ColorSpace) quantizeColor(c color.RGBA, levels int) color.RGBA {
	step := 256 / levels
	return color.RGBA{
		R: uint8((int(c.R) / step) * step),
		G: uint8((int(c.G) / step) * step),
		B: uint8((int(c.B) / step) * step),
		A: c.A,
	}
}
