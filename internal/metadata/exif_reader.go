package metadata

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"github.com/sirupsen/logrus"
)

// EXIFReader extracts EXIF metadata from image files
type EXIFReader struct {
	logger *logrus.Logger
}

// NewEXIFReader creates a new EXIF metadata reader
func NewEXIFReader() *EXIFReader {
	// Register manufacturer notes for better EXIF parsing
	exif.RegisterParsers(mknote.All...)

	return &EXIFReader{
		logger: logrus.New(),
	}
}

// ExtractEXIF extracts EXIF metadata from an image file
func (e *EXIFReader) ExtractEXIF(filePath string) (*api.EXIFInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode EXIF data
	x, err := exif.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode EXIF: %w", err)
	}

	exifInfo := &api.EXIFInfo{}

	// Extract camera information
	if model, err := x.Get(exif.Model); err == nil {
		exifInfo.CameraModel, _ = model.StringVal()
	}

	if make, err := x.Get(exif.Make); err == nil {
		// Combine make and model for better camera identification
		if makeStr, err := make.StringVal(); err == nil {
			if exifInfo.CameraModel != "" {
				exifInfo.CameraModel = makeStr + " " + exifInfo.CameraModel
			} else {
				exifInfo.CameraModel = makeStr
			}
		}
	}

	// Extract lens information
	if lens, err := x.Get(exif.LensModel); err == nil {
		exifInfo.LensModel, _ = lens.StringVal()
	}

	// Extract exposure settings
	if iso, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if isoVal, err := iso.Int(0); err == nil {
			exifInfo.ISO = isoVal
		}
	}

	if exposure, err := x.Get(exif.ExposureTime); err == nil {
		if num, denom, err := exposure.Rat2(0); err == nil {
			if denom != 0 {
				exifInfo.Exposure = fmt.Sprintf("1/%d", denom/num)
			}
		}
	}

	if aperture, err := x.Get(exif.FNumber); err == nil {
		if num, denom, err := aperture.Rat2(0); err == nil {
			if denom != 0 {
				exifInfo.Aperture = float64(num) / float64(denom)
			}
		}
	}

	if focal, err := x.Get(exif.FocalLength); err == nil {
		if num, denom, err := focal.Rat2(0); err == nil {
			if denom != 0 {
				exifInfo.FocalLength = float64(num) / float64(denom)
			}
		}
	}

	// Extract date/time
	if dateTime, err := x.Get(exif.DateTime); err == nil {
		if dateTimeStr, err := dateTime.StringVal(); err == nil {
			if takenAt, err := time.Parse("2006:01:02 15:04:05", dateTimeStr); err == nil {
				exifInfo.TakenAt = takenAt
			}
		}
	}

	// Extract GPS coordinates
	if lat, long, err := x.LatLong(); err == nil {
		exifInfo.GPSLat = lat
		exifInfo.GPSLon = long
		exifInfo.HasGPS = true
	}

	e.logger.Debugf("Extracted EXIF metadata from %s: Camera=%s", filePath, exifInfo.CameraModel)
	return exifInfo, nil
}

// HasEXIFData checks if a file contains EXIF metadata
func (e *EXIFReader) HasEXIFData(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	_, err = exif.Decode(file)
	if err != nil {
		// Handle "no EXIF" case safely across different library versions
		if strings.Contains(err.Error(), "no exif") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetSupportedFormats returns image formats that typically contain EXIF data
func (e *EXIFReader) GetSupportedFormats() []string {
	return []string{".jpg", ".jpeg", ".tiff", ".tif", ".cr2", ".nef", ".arw"}
}
