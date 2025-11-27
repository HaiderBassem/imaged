package api

import "errors"

// Common errors used throughout the image processing library
var (
	ErrImageNotFound      = errors.New("image not found in index")
	ErrInvalidImageFormat = errors.New("unsupported or invalid image format")
	ErrImageTooSmall      = errors.New("image dimensions too small for analysis")
	ErrStorageClosed      = errors.New("storage backend is closed")
	ErrDuplicateOperation = errors.New("duplicate operation detected")
	ErrInvalidThreshold   = errors.New("invalid similarity threshold value")
	ErrImageDecodeFailed  = errors.New("failed to decode image data")
	ErrIndexCorrupted     = errors.New("image index is corrupted")
	ErrInsufficientMemory = errors.New("insufficient memory for operation")
)
