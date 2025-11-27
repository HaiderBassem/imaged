package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"os"
)

// ExactHash computes exact file hashes for duplicate detection
type ExactHash struct {
	hasher hash.Hash
}

// NewExactHash creates a new exact hash calculator
func NewExactHash() *ExactHash {
	return &ExactHash{
		hasher: sha256.New(),
	}
}

// ComputeFileHash calculates the SHA256 hash of a file
func (e *ExactHash) ComputeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	e.hasher.Reset()
	if _, err := io.Copy(e.hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(e.hasher.Sum(nil)), nil
}

// ComputePartialHash calculates hash of first N bytes for quick comparison
func (e *ExactHash) ComputePartialHash(filePath string, bytes int64) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	e.hasher.Reset()
	if _, err := io.CopyN(e.hasher, file, bytes); err != nil && err != io.EOF {
		return "", err
	}

	return hex.EncodeToString(e.hasher.Sum(nil)), nil
}

// CompareHashes compares two hash values for equality
func (e *ExactHash) CompareHashes(hash1, hash2 string) bool {
	return hash1 == hash2
}

// IsValidHash checks if a string is a valid SHA256 hash
func (e *ExactHash) IsValidHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}

	_, err := hex.DecodeString(hash)
	return err == nil
}
