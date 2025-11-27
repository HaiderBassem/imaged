package engine_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/HaiderBassem/imaged/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_ScanFolder(t *testing.T) {
	// Create temporary directory with test images
	tempDir := t.TempDir()
	createTestImages(t, tempDir)

	cfg := engine.DefaultConfig()
	cfg.IndexPath = filepath.Join(tempDir, "test.db")

	eng, err := engine.NewEngine(cfg)
	require.NoError(t, err)
	defer eng.Close()

	ctx := context.Background()
	err = eng.ScanFolder(ctx, tempDir, nil)
	assert.NoError(t, err)

	stats, err := eng.GetStats()
	assert.NoError(t, err)
	assert.Greater(t, stats.TotalImages, int64(0))
}

func TestEngine_FindExactDuplicates(t *testing.T) {
	tempDir := t.TempDir()

	// Create identical files
	data := []byte("fake image data")
	file1 := filepath.Join(tempDir, "image1.jpg")
	file2 := filepath.Join(tempDir, "image2.jpg")

	require.NoError(t, os.WriteFile(file1, data, 0644))
	require.NoError(t, os.WriteFile(file2, data, 0644))

	cfg := engine.DefaultConfig()
	cfg.IndexPath = filepath.Join(tempDir, "test.db")

	eng, err := engine.NewEngine(cfg)
	require.NoError(t, err)
	defer eng.Close()

	ctx := context.Background()
	require.NoError(t, eng.ScanFolder(ctx, tempDir, nil))

	duplicates, err := eng.FindExactDuplicates()
	assert.NoError(t, err)
	assert.Len(t, duplicates, 1)
	assert.Len(t, duplicates[0].DuplicateIDs, 1)
}

func TestEngine_ImageQuality(t *testing.T) {
	tempDir := t.TempDir()
	testImage := createTestImage(t, tempDir, "test.jpg")

	cfg := engine.DefaultConfig()
	eng, err := engine.NewEngine(cfg)
	require.NoError(t, err)
	defer eng.Close()

	quality, err := eng.RateImageQuality(testImage)
	assert.NoError(t, err)
	assert.Greater(t, quality.FinalScore, 0.0)
	assert.LessOrEqual(t, quality.FinalScore, 100.0)
}

// Helper functions
func createTestImages(t *testing.T, dir string) {
	// Create some dummy image files for testing
	extensions := []string{".jpg", ".png", ".webp"}

	for i, ext := range extensions {
		createTestImage(t, dir, fmt.Sprintf("test%d%s", i, ext))
	}
}

func createTestImage(t *testing.T, dir, filename string) string {
	// In real tests, you would create actual image files
	// For now, create empty files
	path := filepath.Join(dir, filename)
	require.NoError(t, os.WriteFile(path, []byte("fake image data"), 0644))
	return path
}
