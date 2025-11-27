package metadata

import (
	"fmt"
	"time"

	"github.com/HaiderBassem/imaged/pkg/api"
)

// CameraInfo represents camera and lens information
type CameraInfo struct {
	Make         string `json:"make"`
	Model        string `json:"model"`
	LensModel    string `json:"lens_model,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
}

// ExposureInfo represents exposure settings
type ExposureInfo struct {
	ISO             int     `json:"iso,omitempty"`
	ExposureTime    string  `json:"exposure_time,omitempty"`
	Aperture        float64 `json:"aperture,omitempty"`
	FocalLength     float64 `json:"focal_length,omitempty"`
	ExposureProgram int     `json:"exposure_program,omitempty"`
	ExposureBias    float64 `json:"exposure_bias,omitempty"`
	MeteringMode    int     `json:"metering_mode,omitempty"`
	Flash           int     `json:"flash,omitempty"`
}

// GPSInfo represents GPS coordinates and location data
type GPSInfo struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Altitude  float64   `json:"altitude,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// ImageProperties represents additional image properties
type ImageProperties struct {
	Orientation    int    `json:"orientation,omitempty"`
	ColorSpace     int    `json:"color_space,omitempty"`
	XResolution    int    `json:"x_resolution,omitempty"`
	YResolution    int    `json:"y_resolution,omitempty"`
	ResolutionUnit int    `json:"resolution_unit,omitempty"`
	Software       string `json:"software,omitempty"`
	Artist         string `json:"artist,omitempty"`
	Copyright      string `json:"copyright,omitempty"`
}

// ThumbnailInfo represents embedded thumbnail information
type ThumbnailInfo struct {
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Format string `json:"format,omitempty"`
	Size   int    `json:"size,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// CompleteMetadata represents comprehensive image metadata
type CompleteMetadata struct {
	// Basic file information
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	FileFormat string    `json:"file_format"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
	AccessedAt time.Time `json:"accessed_at"`

	// Image dimensions and properties
	Width      int              `json:"width"`
	Height     int              `json:"height"`
	Properties *ImageProperties `json:"properties,omitempty"`

	// Camera and exposure information
	Camera   *CameraInfo   `json:"camera,omitempty"`
	Exposure *ExposureInfo `json:"exposure,omitempty"`

	// Location information
	GPS *GPSInfo `json:"gps,omitempty"`

	// Thumbnail information
	Thumbnail *ThumbnailInfo `json:"thumbnail,omitempty"`

	// Processing information
	ProcessedAt time.Time `json:"processed_at"`
	Version     string    `json:"version"`
}

// NewCompleteMetadata creates a new CompleteMetadata instance
func NewCompleteMetadata() *CompleteMetadata {
	return &CompleteMetadata{
		ProcessedAt: time.Now(),
		Version:     "1.0",
	}
}

// HasCameraInfo checks if camera information is available
func (m *CompleteMetadata) HasCameraInfo() bool {
	return m.Camera != nil && m.Camera.Model != ""
}

// HasExposureInfo checks if exposure information is available
func (m *CompleteMetadata) HasExposureInfo() bool {
	return m.Exposure != nil && (m.Exposure.ISO > 0 || m.Exposure.ExposureTime != "")
}

// HasGPSInfo checks if GPS information is available
func (m *CompleteMetadata) HasGPSInfo() bool {
	return m.GPS != nil && m.GPS.Latitude != 0 && m.GPS.Longitude != 0
}

// GetAspectRatio returns the image aspect ratio
func (m *CompleteMetadata) GetAspectRatio() float64 {
	if m.Height == 0 {
		return 0
	}
	return float64(m.Width) / float64(m.Height)
}

// IsPortrait returns true if image is in portrait orientation
func (m *CompleteMetadata) IsPortrait() bool {
	return m.Height > m.Width
}

// IsLandscape returns true if image is in landscape orientation
func (m *CompleteMetadata) IsLandscape() bool {
	return m.Width > m.Height
}

// GetMegapixels returns the image resolution in megapixels
func (m *CompleteMetadata) GetMegapixels() float64 {
	return float64(m.Width*m.Height) / 1000000.0
}

// Validate checks if the metadata is valid
func (m *CompleteMetadata) Validate() error {
	if m.FilePath == "" {
		return fmt.Errorf("missing file path")
	}

	if m.FileSize <= 0 {
		return fmt.Errorf("invalid file size: %d", m.FileSize)
	}

	if m.Width <= 0 || m.Height <= 0 {
		return fmt.Errorf("invalid dimensions: %dx%d", m.Width, m.Height)
	}

	if m.FileFormat == "" {
		return fmt.Errorf("missing file format")
	}

	return nil
}

// ToBasicMetadata converts to the basic API metadata format
func (m *CompleteMetadata) ToBasicMetadata() *api.ImageMetadata {
	basic := &api.ImageMetadata{
		Path:       m.FilePath,
		SizeBytes:  m.FileSize,
		Format:     m.FileFormat,
		Width:      m.Width,
		Height:     m.Height,
		ModifiedAt: m.ModifiedAt,
	}

	if m.Camera != nil || m.Exposure != nil || m.GPS != nil {
		basic.EXIF = &api.EXIFInfo{}

		if m.Camera != nil {
			basic.EXIF.CameraModel = m.Camera.Model
			basic.EXIF.LensModel = m.Camera.LensModel
		}

		if m.Exposure != nil {
			basic.EXIF.ISO = m.Exposure.ISO
			basic.EXIF.Exposure = m.Exposure.ExposureTime
			basic.EXIF.Aperture = m.Exposure.Aperture
			basic.EXIF.FocalLength = m.Exposure.FocalLength
		}

		if m.GPS != nil {
			basic.EXIF.GPSLat = m.GPS.Latitude
			basic.EXIF.GPSLon = m.GPS.Longitude
			basic.EXIF.HasGPS = true
		}
	}

	return basic
}
