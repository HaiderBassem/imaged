package commands

import (
	"fmt"
	"os"

	"github.com/HaiderBassem/imaged/pkg/engine"
	"github.com/urfave/cli/v2"
)

// QualityCommand handles image quality analysis
func QualityCommand(c *cli.Context) error {
	imagePath := c.String("image")

	if imagePath == "" {
		return cli.Exit("Image path is required", 1)
	}

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return cli.Exit("Image file does not exist", 1)
	}

	fmt.Printf("Analyzing image quality: %s\n", imagePath)

	cfg := engine.DefaultConfig()
	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create engine: %v", err), 1)
	}
	defer eng.Close()

	quality, err := eng.RateImageQuality(imagePath)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to analyze quality: %v", err), 1)
	}

	// Display quality metrics
	fmt.Printf("\nQuality Analysis Results:\n")
	fmt.Printf("  Overall Score: %.1f/100\n", quality.FinalScore)
	fmt.Printf("  Sharpness: %.3f\n", quality.Sharpness)
	fmt.Printf("  Noise: %.3f\n", quality.Noise)
	fmt.Printf("  Exposure: %.3f\n", quality.Exposure)
	fmt.Printf("  Contrast: %.3f\n", quality.Contrast)
	fmt.Printf("  Compression: %.3f\n", quality.Compression)
	fmt.Printf("  Color Cast: %.3f\n", quality.ColorCast)

	// Provide recommendations
	fmt.Printf("\nRecommendations:\n")
	if quality.Sharpness < 0.3 {
		fmt.Printf("    Image is blurry (sharpness: %.2f)\n", quality.Sharpness)
	}
	if quality.Noise > 0.7 {
		fmt.Printf("    High noise level (%.2f)\n", quality.Noise)
	}
	if abs(quality.Exposure-0.5) > 0.3 {
		status := "overexposed"
		if quality.Exposure < 0.5 {
			status = "underexposed"
		}
		fmt.Printf("    Image is %s (exposure: %.2f)\n", status, quality.Exposure)
	}
	if quality.FinalScore >= 80 {
		fmt.Printf("   Excellent quality image\n")
	} else if quality.FinalScore >= 60 {
		fmt.Printf("    Average quality image\n")
	} else {
		fmt.Printf("   Poor quality image\n")
	}

	return nil
}

// abs returns absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
