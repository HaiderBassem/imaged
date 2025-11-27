package similarity

import "math"

// Distance provides various distance calculation methods
type Distance struct{}

// NewDistance creates a new distance calculator
func NewDistance() *Distance {
	return &Distance{}
}

// EuclideanDistance calculates Euclidean distance between two vectors
func (d *Distance) EuclideanDistance(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return math.MaxFloat64
	}

	var sum float64
	for i := range vec1 {
		diff := vec1[i] - vec2[i]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// CosineSimilarity calculates cosine similarity between two vectors
func (d *Distance) CosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct, mag1, mag2 float64
	for i := range vec1 {
		dotProduct += vec1[i] * vec2[i]
		mag1 += vec1[i] * vec1[i]
		mag2 += vec2[i] * vec2[i]
	}

	if mag1 == 0 || mag2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(mag1) * math.Sqrt(mag2))
}

// ManhattanDistance calculates Manhattan distance between two vectors
func (d *Distance) ManhattanDistance(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return math.MaxFloat64
	}

	var sum float64
	for i := range vec1 {
		sum += math.Abs(vec1[i] - vec2[i])
	}

	return sum
}

// HammingDistance calculates Hamming distance between two integers
func (d *Distance) HammingDistance(a, b uint64) int {
	xor := a ^ b
	distance := 0
	for xor != 0 {
		distance++
		xor &= xor - 1
	}
	return distance
}

// ChiSquaredDistance calculates Chi-squared distance between two histograms
func (d *Distance) ChiSquaredDistance(hist1, hist2 []float64) float64 {
	if len(hist1) != len(hist2) {
		return math.MaxFloat64
	}

	var sum float64
	for i := range hist1 {
		if hist1[i]+hist2[i] > 0 {
			diff := hist1[i] - hist2[i]
			sum += (diff * diff) / (hist1[i] + hist2[i])
		}
	}

	return sum / 2.0
}
