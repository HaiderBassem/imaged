package imaging

import (
	"fmt"
	"image"
	"io"
	"os"

	"github.com/dustin/go-humanize"
)

// Decoder handles image decoding with format detection and error handling
type Decoder struct {
	supportedFormats map[string]bool
}

// NewDecoder creates a new image decoder
func NewDecoder() *Decoder {
	return &Decoder{
		supportedFormats: map[string]bool{
			"jpeg": true, "jpg": true, "png": true,
			"gif": true, "bmp": true, "tiff": true,
			"webp": true,
		},
	}
}

// DecodeImage decodes an image from a file with comprehensive error handling
func (d *Decoder) DecodeImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Decode image with format detection
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Verify supported format
	if !d.supportedFormats[format] {
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	return img, nil
}

// DecodeImageConfig decodes only image configuration (dimensions, format) without loading full image
func (d *Decoder) DecodeImageConfig(filePath string) (image.Config, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return image.Config{}, "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return image.Config{}, "", fmt.Errorf("failed to decode image config: %w", err)
	}

	if !d.supportedFormats[format] {
		return image.Config{}, "", fmt.Errorf("unsupported image format: %s", format)
	}

	return config, format, nil
}

// DecodeWithMetadata decodes image and returns additional metadata
func (d *Decoder) DecodeWithMetadata(filePath string) (*ImageWithMetadata, error) {
	config, format, err := d.DecodeImageConfig(filePath)
	if err != nil {
		return nil, err
	}

	img, err := d.DecodeImage(filePath)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &ImageWithMetadata{
		Image:    img,
		Config:   config,
		Format:   format,
		FileSize: fileInfo.Size(),
		FilePath: filePath,
	}, nil
}

// ImageWithMetadata contains image data and associated metadata
type ImageWithMetadata struct {
	Image    image.Image
	Config   image.Config
	Format   string
	FileSize int64
	FilePath string
}

// GetDimensions returns image dimensions
func (im *ImageWithMetadata) GetDimensions() (int, int) {
	return im.Config.Width, im.Config.Height
}

// GetAspectRatio returns image aspect ratio
func (im *ImageWithMetadata) GetAspectRatio() float64 {
	if im.Config.Height == 0 {
		return 0
	}
	return float64(im.Config.Width) / float64(im.Config.Height)
}

// IsPortrait returns true if image is in portrait orientation
func (im *ImageWithMetadata) IsPortrait() bool {
	return im.Config.Height > im.Config.Width
}

// IsLandscape returns true if image is in landscape orientation
func (im *ImageWithMetadata) IsLandscape() bool {
	return im.Config.Width > im.Config.Height
}

// ValidateImage performs basic image validation
func (d *Decoder) ValidateImage(filePath string) error {
	config, format, err := d.DecodeImageConfig(filePath)
	if err != nil {
		return err
	}

	// Check dimensions
	if config.Width <= 0 || config.Height <= 0 {
		return fmt.Errorf("invalid image dimensions: %dx%d", config.Width, config.Height)
	}

	// Check if image is too large (prevent memory issues)
	maxPixels := 100000000 // 100MP
	if config.Width*config.Height > maxPixels {
		return fmt.Errorf("image too large: %dx%d pixels", config.Width, config.Height)
	}

	// Check file size (rough estimation)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	maxFileSize := int64(500 * 1024 * 1024) // 500MB
	if fileInfo.Size() > maxFileSize {
		return fmt.Errorf("file too large: %s", humanize.Bytes(uint64(fileInfo.Size())))
	}

	// Validate format-specific constraints
	switch format {
	case "jpeg", "jpg":
		// JPEG specific validations
		if config.Width < 16 || config.Height < 16 {
			return fmt.Errorf("JPEG image too small: %dx%d", config.Width, config.Height)
		}
	case "png":
		// PNG specific validations
		if config.Width > 32767 || config.Height > 32767 {
			return fmt.Errorf("PNG dimensions too large: %dx%d", config.Width, config.Height)
		}
	}

	return nil
}

// GetSupportedFormats returns list of supported image formats
func (d *Decoder) GetSupportedFormats() []string {
	formats := make([]string, 0, len(d.supportedFormats))
	for format := range d.supportedFormats {
		formats = append(formats, format)
	}
	return formats
}

// IsFormatSupported checks if a format is supported
func (d *Decoder) IsFormatSupported(format string) bool {
	return d.supportedFormats[format]
}

// DecodeFromReader decodes image from an io.Reader
func (d *Decoder) DecodeFromReader(reader io.Reader) (image.Image, string, error) {
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image from reader: %w", err)
	}

	if !d.supportedFormats[format] {
		return nil, "", fmt.Errorf("unsupported image format: %s", format)
	}

	return img, format, nil
}
