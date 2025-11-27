package imaging

import (
	"image"
	"math"

	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
)

// Transformer provides image transformation utilities
type Transformer struct {
	maxDimension int
}

// NewTransformer creates a new image transformer
func NewTransformer(maxDimension int) *Transformer {
	return &Transformer{
		maxDimension: maxDimension,
	}
}

// Resize resizes image while maintaining aspect ratio
func (t *Transformer) Resize(img image.Image, width, height int) image.Image {
	return resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
}

// ResizeToFit resizes image to fit within specified dimensions while maintaining aspect ratio
func (t *Transformer) ResizeToFit(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Calculate new dimensions maintaining aspect ratio
	newWidth, newHeight := origWidth, origHeight

	if origWidth > maxWidth {
		newWidth = maxWidth
		newHeight = int(float64(origHeight) * float64(maxWidth) / float64(origWidth))
	}

	if newHeight > maxHeight {
		newHeight = maxHeight
		newWidth = int(float64(origWidth) * float64(maxHeight) / float64(origHeight))
	}

	return t.Resize(img, newWidth, newHeight)
}

// ResizeToFill resizes and crops image to exactly fill specified dimensions
func (t *Transformer) ResizeToFill(img image.Image, width, height int) image.Image {
	// First resize to cover the area
	resized := t.ResizeToCover(img, width, height)

	// Then crop to exact dimensions
	return imaging.CropCenter(resized, width, height)
}

// ResizeToCover resizes image to cover specified area while maintaining aspect ratio
func (t *Transformer) ResizeToCover(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Calculate scale ratio
	scaleX := float64(width) / float64(origWidth)
	scaleY := float64(height) / float64(origHeight)
	scale := math.Max(scaleX, scaleY)

	newWidth := int(float64(origWidth) * scale)
	newHeight := int(float64(origHeight) * scale)

	return t.Resize(img, newWidth, newHeight)
}

// Crop crops image to specified rectangle
func (t *Transformer) Crop(img image.Image, x, y, width, height int) image.Image {
	return imaging.Crop(img, image.Rect(x, y, x+width, y+height))
}

// CropCenter crops image from center
func (t *Transformer) CropCenter(img image.Image, width, height int) image.Image {
	return imaging.CropCenter(img, width, height)
}

// Rotate rotates image by specified angle (in degrees)
func (t *Transformer) Rotate(img image.Image, angle float64) image.Image {
	return imaging.Rotate(img, angle, image.Transparent)
}

// FlipHorizontal flips image horizontally
func (t *Transformer) FlipHorizontal(img image.Image) image.Image {
	return imaging.FlipH(img)
}

// FlipVertical flips image vertically
func (t *Transformer) FlipVertical(img image.Image) image.Image {
	return imaging.FlipV(img)
}

// Grayscale converts image to grayscale
func (t *Transformer) Grayscale(img image.Image) image.Image {
	return imaging.Grayscale(img)
}

// AdjustBrightness adjusts image brightness
func (t *Transformer) AdjustBrightness(img image.Image, percentage float64) image.Image {
	return imaging.AdjustBrightness(img, percentage)
}

// AdjustContrast adjusts image contrast
func (t *Transformer) AdjustContrast(img image.Image, percentage float64) image.Image {
	return imaging.AdjustContrast(img, percentage)
}

// AdjustSaturation adjusts image saturation
func (t *Transformer) AdjustSaturation(img image.Image, percentage float64) image.Image {
	return imaging.AdjustSaturation(img, percentage)
}

// Blur applies Gaussian blur to image
func (t *Transformer) Blur(img image.Image, sigma float64) image.Image {
	return imaging.Blur(img, sigma)
}

// Sharpen applies sharpening filter to image
func (t *Transformer) Sharpen(img image.Image, sigma float64) image.Image {
	return imaging.Sharpen(img, sigma)
}

// CreateThumbnail creates a thumbnail image
func (t *Transformer) CreateThumbnail(img image.Image, size int) image.Image {
	return t.ResizeToFit(img, size, size)
}

// ExtractRegion extracts a region of interest from image
func (t *Transformer) ExtractRegion(img image.Image, region image.Rectangle) image.Image {
	bounds := img.Bounds()

	// Ensure region is within bounds
	region = region.Intersect(bounds)

	return imaging.Crop(img, region)
}

// CreateCollage creates a collage from multiple images
func (t *Transformer) CreateCollage(images []image.Image, cols int, spacing int) image.Image {
	if len(images) == 0 {
		return nil
	}

	// Calculate dimensions for collage
	rows := int(math.Ceil(float64(len(images)) / float64(cols)))

	// Find maximum dimensions for each cell
	var maxCellWidth, maxCellHeight int
	for _, img := range images {
		bounds := img.Bounds()
		if bounds.Dx() > maxCellWidth {
			maxCellWidth = bounds.Dx()
		}
		if bounds.Dy() > maxCellHeight {
			maxCellHeight = bounds.Dy()
		}
	}

	// Create collage image
	collageWidth := cols*maxCellWidth + (cols-1)*spacing
	collageHeight := rows*maxCellHeight + (rows-1)*spacing
	collage := image.NewRGBA(image.Rect(0, 0, collageWidth, collageHeight))

	// Place images in collage
	for i, img := range images {
		row := i / cols
		col := i % cols

		x := col * (maxCellWidth + spacing)
		y := row * (maxCellHeight + spacing)

		// Resize image to fit cell
		resized := t.ResizeToFit(img, maxCellWidth, maxCellHeight)

		// Calculate position to center image in cell
		bounds := resized.Bounds()
		offsetX := x + (maxCellWidth-bounds.Dx())/2
		offsetY := y + (maxCellHeight-bounds.Dy())/2

		// Draw image on collage
		imaging.Paste(collage, resized, image.Pt(offsetX, offsetY))
	}

	return collage
}

// NormalizeOrientation corrects image orientation based on EXIF data
func (t *Transformer) NormalizeOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Transpose(img)
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Transverse(img)
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}

// GetTransformationInfo returns information about image transformations
func (t *Transformer) GetTransformationInfo(original, transformed image.Image) map[string]interface{} {
	origBounds := original.Bounds()
	transBounds := transformed.Bounds()

	info := make(map[string]interface{})
	info["original_width"] = origBounds.Dx()
	info["original_height"] = origBounds.Dy()
	info["transformed_width"] = transBounds.Dx()
	info["transformed_height"] = transBounds.Dy()
	info["scale_x"] = float64(transBounds.Dx()) / float64(origBounds.Dx())
	info["scale_y"] = float64(transBounds.Dy()) / float64(origBounds.Dy())
	info["aspect_ratio_original"] = float64(origBounds.Dx()) / float64(origBounds.Dy())
	info["aspect_ratio_transformed"] = float64(transBounds.Dx()) / float64(transBounds.Dy())

	return info
}
