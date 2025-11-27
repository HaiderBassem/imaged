package quality

import (
	"image"
	"math"

	"github.com/disintegration/imaging"
)

// NoiseAnalyzer estimates image noise levels
type NoiseAnalyzer struct{}

// NewNoiseAnalyzer creates a new noise analyzer
func NewNoiseAnalyzer() *NoiseAnalyzer {
	return &NoiseAnalyzer{}
}

// helper: convert pixel to grayscale value
func grayValue(img image.Image, x, y int) float64 {
	r, g, b, _ := img.At(x, y).RGBA()
	return float64(r+g+b) / 3.0
}

// AnalyzeNoise estimates image noise level
func (n *NoiseAnalyzer) AnalyzeNoise(img image.Image) (float64, error) {
	gray := imaging.Grayscale(img)
	bounds := gray.Bounds()

	if bounds.Dx() < 3 || bounds.Dy() < 3 {
		return 0.5, nil
	}

	var noiseSum float64
	var count int

	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {

			center := grayValue(gray, x, y)

			neighbors := []float64{
				grayValue(gray, x-1, y-1),
				grayValue(gray, x, y-1),
				grayValue(gray, x+1, y-1),
				grayValue(gray, x-1, y),
				grayValue(gray, x+1, y),
				grayValue(gray, x-1, y+1),
				grayValue(gray, x, y+1),
				grayValue(gray, x+1, y+1),
			}

			var mean, variance float64

			for _, n := range neighbors {
				mean += n
			}
			mean /= float64(len(neighbors))

			for _, n := range neighbors {
				diff := n - mean
				variance += diff * diff
			}
			variance /= float64(len(neighbors))

			gradient := math.Abs(center - mean)
			if gradient < 10 {
				noiseSum += math.Sqrt(variance)
				count++
			}
		}
	}

	if count == 0 {
		return 0.5, nil
	}

	avgNoise := noiseSum / float64(count)
	return math.Min(avgNoise/50.0, 1.0), nil
}

// DetectNoisePattern identifies specific noise patterns
func (n *NoiseAnalyzer) DetectNoisePattern(img image.Image) map[string]float64 {
	patterns := make(map[string]float64)
	gray := imaging.Grayscale(img)

	patterns["high_frequency"] = n.analyzeHighFrequencyNoise(gray)
	patterns["low_frequency"] = n.analyzeLowFrequencyNoise(gray)
	patterns["salt_pepper"] = n.analyzeSaltPepperNoise(gray)

	return patterns
}

func (n *NoiseAnalyzer) analyzeHighFrequencyNoise(gray image.Image) float64 {
	bounds := gray.Bounds()
	var hfNoise float64
	var samples int

	for y := 1; y < bounds.Max.Y-1; y++ {
		for x := 1; x < bounds.Max.X-1; x++ {
			center := grayValue(gray, x, y)

			var neighborSum float64
			var count int

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					neighborSum += grayValue(gray, x+dx, y+dy)
					count++
				}
			}

			avg := neighborSum / float64(count)
			if math.Abs(center-avg) > 5 {
				hfNoise++
				samples++
			}
		}
	}

	if samples == 0 {
		return 0
	}

	return math.Min(hfNoise/float64(samples), 1.0)
}

func (n *NoiseAnalyzer) analyzeLowFrequencyNoise(gray image.Image) float64 {
	bounds := gray.Bounds()
	var lfNoise float64
	var samples int

	kernel := 5
	half := kernel / 2

	for y := half; y < bounds.Max.Y-half; y += 2 {
		for x := half; x < bounds.Max.X-half; x += 2 {
			var sum float64
			var count int

			for dy := -half; dy <= half; dy++ {
				for dx := -half; dx <= half; dx++ {
					sum += grayValue(gray, x+dx, y+dy)
					count++
				}
			}

			mean := sum / float64(count)
			center := grayValue(gray, x, y)

			if math.Abs(center-mean) > 2 {
				lfNoise++
				samples++
			}
		}
	}

	if samples == 0 {
		return 0
	}

	return math.Min(lfNoise/float64(samples), 1.0)
}

func (n *NoiseAnalyzer) analyzeSaltPepperNoise(gray image.Image) float64 {
	bounds := gray.Bounds()
	var extreme int
	total := bounds.Dx() * bounds.Dy()

	for y := 1; y < bounds.Max.Y-1; y++ {
		for x := 1; x < bounds.Max.X-1; x++ {
			center := grayValue(gray, x, y)

			var sum float64
			var count int

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					sum += grayValue(gray, x+dx, y+dy)
					count++
				}
			}

			avg := sum / float64(count)

			if math.Abs(center-avg) > 50 {
				extreme++
			}
		}
	}

	return math.Min(float64(extreme)/float64(total)*10.0, 1.0)
}
