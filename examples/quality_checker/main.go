package main

import (
	"fmt"
	"log"

	"github.com/HaiderBassem/imaged/pkg/engine"
)

func main() {
	eng, err := engine.NewEngine(engine.DefaultConfig())
	if err != nil {
		log.Fatal("Failed to create engine:", err)
	}
	defer eng.Close()

	// Analyze image quality
	quality, err := eng.RateImageQuality("test_image.jpg")
	if err != nil {
		log.Fatal("Failed to analyze quality:", err)
	}

	fmt.Printf("Quality Analysis:\n")
	fmt.Printf("  Overall Score: %.1f/100\n", quality.FinalScore)
	fmt.Printf("  Sharpness: %.3f\n", quality.Sharpness)
	fmt.Printf("  Noise: %.3f\n", quality.Noise)
	fmt.Printf("  Exposure: %.3f\n", quality.Exposure)

	if quality.FinalScore > 80 {
		fmt.Println("✅ High quality image")
	} else if quality.FinalScore > 60 {
		fmt.Println("⚠️  Medium quality image")
	} else {
		fmt.Println("❌ Low quality image")
	}
}
