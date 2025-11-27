package hash

import (
	"image"
	"math"

	"github.com/disintegration/imaging"
)

// AdvancedHash provides advanced hashing algorithms beyond basic perceptual hashes
type AdvancedHash struct {
	colorSig *ColorSignature
}

// NewAdvancedHash creates a new advanced hash calculator
func NewAdvancedHash() *AdvancedHash {
	return &AdvancedHash{
		colorSig: NewColorSignature(16),
	}
}

// ComputeFeatureVector computes a comprehensive feature vector for an image
func (a *AdvancedHash) ComputeFeatureVector(img image.Image) ([]float64, error) {
	var features []float64

	// Color features
	colorHist, err := a.colorSig.ComputeColorHistogram(img)
	if err != nil {
		return nil, err
	}
	features = append(features, colorHist...)

	// Texture features (simplified)
	textureFeatures := a.computeTextureFeatures(img)
	features = append(features, textureFeatures...)

	// Shape features (simplified)
	shapeFeatures := a.computeShapeFeatures(img)
	features = append(features, shapeFeatures...)

	return features, nil
}

// computeTextureFeatures calculates basic texture features
func (a *AdvancedHash) computeTextureFeatures(img image.Image) []float64 {
	// Simplified texture analysis using edge density
	bounds := img.Bounds()
	gray := imaging.Grayscale(img)

	var edgePixels int
	totalPixels := bounds.Dx() * bounds.Dy()

	// Simple edge detection using gradient magnitude
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			cr, _, _, _ := gray.At(x, y).RGBA()
			rr, _, _, _ := gray.At(x+1, y).RGBA()
			br, _, _, _ := gray.At(x, y+1).RGBA()

			gradX := math.Abs(float64(cr) - float64(rr))
			gradY := math.Abs(float64(cr) - float64(br))

			// Threshold for edge presence
			if gradX > 10000 || gradY > 10000 {
				edgePixels++
			}
		}
	}

	edgeDensity := 0.0
	if totalPixels > 0 {
		edgeDensity = float64(edgePixels) / float64(totalPixels)
	}

	return []float64{edgeDensity}
}

// computeShapeFeatures calculates basic shape features
func (a *AdvancedHash) computeShapeFeatures(img image.Image) []float64 {
	bounds := img.Bounds()

	// Aspect ratio
	aspectRatio := float64(bounds.Dx()) / float64(bounds.Dy())

	// Centroid (simplified)
	centroidX := float64(bounds.Dx()) / 2
	centroidY := float64(bounds.Dy()) / 2

	return []float64{aspectRatio, centroidX, centroidY}
}

// ComputeDeepHash computes a deep feature-based hash
func (a *AdvancedHash) ComputeDeepHash(img image.Image) ([]float32, error) {
	featureVector, err := a.ComputeFeatureVector(img)
	if err != nil {
		return nil, err
	}

	// Convert to float32 for compatibility with ML models
	deepHash := make([]float32, len(featureVector))
	for i, val := range featureVector {
		deepHash[i] = float32(val)
	}

	return deepHash, nil
}

// CompareFeatureVectors compares two feature vectors using cosine similarity
func (a *AdvancedHash) CompareFeatureVectors(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}

	var dotProduct, mag1, mag2 float64
	for i := range vec1 {
		dotProduct += vec1[i] * vec2[i]
		mag1 += vec1[i] * vec1[i]
		mag2 += vec2[i] * vec2[i]
	}

	if mag1 == 0 || mag2 == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(mag1) * math.Sqrt(mag2))
}
