package quality

import "github.com/HaiderBassem/imaged/pkg/api"

// ScoreCalculator calculates final quality scores from individual metrics
type ScoreCalculator struct {
	weights map[string]float64
}

// NewScoreCalculator creates a new score calculator with default weights
func NewScoreCalculator() *ScoreCalculator {
	return &ScoreCalculator{
		weights: map[string]float64{
			"sharpness":   0.30,
			"noise":       0.25,
			"exposure":    0.20,
			"contrast":    0.15,
			"compression": 0.05,
			"color_cast":  0.05,
		},
	}
}

// CalculateFinalScore computes the overall quality score
func (sc *ScoreCalculator) CalculateFinalScore(quality *api.ImageQuality) float64 {
	// Adjust exposure score to penalize both under and overexposure
	exposureScore := 1.0 - abs(quality.Exposure-0.5)*2

	// Calculate weighted score (invert noise, compression, and color_cast since lower is better)
	score := quality.Sharpness*sc.weights["sharpness"] +
		(1-quality.Noise)*sc.weights["noise"] +
		exposureScore*sc.weights["exposure"] +
		quality.Contrast*sc.weights["contrast"] +
		(1-quality.Compression)*sc.weights["compression"] +
		(1-quality.ColorCast)*sc.weights["color_cast"]

	// Convert to 0-100 scale
	finalScore := score * 100

	// Ensure valid range
	return max(0, min(100, finalScore))
}

// SetWeights allows customizing the weight of each quality metric
func (sc *ScoreCalculator) SetWeights(weights map[string]float64) {
	sc.weights = weights
}

// GetWeights returns the current weight configuration
func (sc *ScoreCalculator) GetWeights() map[string]float64 {
	return sc.weights
}

// NormalizeWeights ensures weights sum to 1.0
func (sc *ScoreCalculator) NormalizeWeights() {
	var total float64
	for _, weight := range sc.weights {
		total += weight
	}

	if total == 0 {
		return
	}

	for key := range sc.weights {
		sc.weights[key] /= total
	}
}

// CalculateSubscores returns individual component scores
func (sc *ScoreCalculator) CalculateSubscores(quality *api.ImageQuality) map[string]float64 {
	exposureScore := 1.0 - abs(quality.Exposure-0.5)*2

	return map[string]float64{
		"sharpness":   quality.Sharpness * 100,
		"noise":       (1 - quality.Noise) * 100,
		"exposure":    exposureScore * 100,
		"contrast":    quality.Contrast * 100,
		"compression": (1 - quality.Compression) * 100,
		"color_cast":  (1 - quality.ColorCast) * 100,
	}
}

// Utility functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
