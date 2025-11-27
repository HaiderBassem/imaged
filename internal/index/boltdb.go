package index

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/HaiderBassem/imaged/pkg/api"
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
)

// BoltStore implements the index Store interface using BoltDB for persistent storage
type BoltStore struct {
	db     *bolt.DB
	logger *logrus.Logger
}

// NewBoltStore creates a new BoltDB-based index store
func NewBoltStore(dbPath string) (*BoltStore, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	store := &BoltStore{
		db:     db,
		logger: logger,
	}

	// Initialize required buckets
	if err := store.initBuckets(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize buckets: %w", err)
	}

	return store, nil
}

// initBuckets creates all necessary buckets if they don't exist
func (s *BoltStore) initBuckets() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		buckets := []string{
			"fingerprints",
			"sha256_index",
			"ahash_index",
			"phash_index",
			"dhash_index",
			"whash_index",
			"path_index",
			"metadata",
		}

		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}
		return nil
	})
}

// SaveFingerprint stores an image fingerprint and updates all indices
func (s *BoltStore) SaveFingerprint(fp api.ImageFingerprint) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Serialize fingerprint data
		data, err := json.Marshal(fp)
		if err != nil {
			return fmt.Errorf("failed to marshal fingerprint: %w", err)
		}

		// Store in main fingerprints bucket
		fingerprintsBucket := tx.Bucket([]byte("fingerprints"))
		if err := fingerprintsBucket.Put([]byte(fp.ID), data); err != nil {
			return fmt.Errorf("failed to store fingerprint: %w", err)
		}

		// Update SHA256 index for exact duplicate detection
		sha256Bucket := tx.Bucket([]byte("sha256_index"))
		if err := sha256Bucket.Put([]byte(fp.Metadata.SHA256), []byte(fp.ID)); err != nil {
			return fmt.Errorf("failed to update SHA256 index: %w", err)
		}

		// Update path index for quick path-based lookups
		pathBucket := tx.Bucket([]byte("path_index"))
		if err := pathBucket.Put([]byte(fp.Metadata.Path), []byte(fp.ID)); err != nil {
			return fmt.Errorf("failed to update path index: %w", err)
		}

		// Update perceptual hash indices
		if err := s.updateHashIndex(tx, "ahash_index", fp.PHashes.AHash, fp.ID); err != nil {
			return err
		}
		if err := s.updateHashIndex(tx, "phash_index", fp.PHashes.PHash, fp.ID); err != nil {
			return err
		}
		if err := s.updateHashIndex(tx, "dhash_index", fp.PHashes.DHash, fp.ID); err != nil {
			return err
		}
		if err := s.updateHashIndex(tx, "whash_index", fp.PHashes.WHash, fp.ID); err != nil {
			return err
		}

		s.logger.Debugf("Successfully indexed image: %s", fp.ID)
		return nil
	})
}

// updateHashIndex updates a specific perceptual hash index
func (s *BoltStore) updateHashIndex(tx *bolt.Tx, bucketName string, hash uint64, imageID api.ImageID) error {
	if hash == 0 {
		return nil // Skip if hash wasn't computed
	}

	bucket := tx.Bucket([]byte(bucketName))
	key := fmt.Sprintf("%016x", hash)

	// Get existing images for this hash
	var images []api.ImageID
	existingData := bucket.Get([]byte(key))
	if existingData != nil {
		if err := json.Unmarshal(existingData, &images); err != nil {
			return fmt.Errorf("failed to unmarshal existing images: %w", err)
		}
	}

	// Add new image ID if not already present
	found := false
	for _, id := range images {
		if id == imageID {
			found = true
			break
		}
	}

	if !found {
		images = append(images, imageID)
		newData, err := json.Marshal(images)
		if err != nil {
			return fmt.Errorf("failed to marshal image list: %w", err)
		}

		if err := bucket.Put([]byte(key), newData); err != nil {
			return fmt.Errorf("failed to update hash index: %w", err)
		}
	}

	return nil
}

// GetFingerprint retrieves a fingerprint by image ID
func (s *BoltStore) GetFingerprint(imageID api.ImageID) (*api.ImageFingerprint, error) {
	var fingerprint api.ImageFingerprint

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("fingerprints"))
		data := bucket.Get([]byte(imageID))
		if data == nil {
			return api.ErrImageNotFound
		}

		if err := json.Unmarshal(data, &fingerprint); err != nil {
			return fmt.Errorf("failed to unmarshal fingerprint: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &fingerprint, nil
}

// GetAllFingerprints retrieves all fingerprints from the index
func (s *BoltStore) GetAllFingerprints() ([]api.ImageFingerprint, error) {
	var fingerprints []api.ImageFingerprint

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("fingerprints"))

		return bucket.ForEach(func(k, v []byte) error {
			var fp api.ImageFingerprint
			if err := json.Unmarshal(v, &fp); err != nil {
				s.logger.Warnf("Failed to unmarshal fingerprint %s: %v", k, err)
				return nil // Continue with next fingerprint
			}
			fingerprints = append(fingerprints, fp)
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve fingerprints: %w", err)
	}

	return fingerprints, nil
}

// DeleteFingerprint removes a fingerprint and all its indices
func (s *BoltStore) DeleteFingerprint(imageID api.ImageID) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get the fingerprint first to update indices
		fp, err := s.GetFingerprint(imageID)
		if err != nil {
			return err
		}

		// Remove from main fingerprints bucket
		fingerprintsBucket := tx.Bucket([]byte("fingerprints"))
		if err := fingerprintsBucket.Delete([]byte(imageID)); err != nil {
			return fmt.Errorf("failed to delete fingerprint: %w", err)
		}

		// Remove from SHA256 index
		sha256Bucket := tx.Bucket([]byte("sha256_index"))
		if err := sha256Bucket.Delete([]byte(fp.Metadata.SHA256)); err != nil {
			return fmt.Errorf("failed to remove SHA256 index: %w", err)
		}

		// Remove from path index
		pathBucket := tx.Bucket([]byte("path_index"))
		if err := pathBucket.Delete([]byte(fp.Metadata.Path)); err != nil {
			return fmt.Errorf("failed to remove path index: %w", err)
		}

		// Remove from perceptual hash indices
		if err := s.removeFromHashIndex(tx, "ahash_index", fp.PHashes.AHash, imageID); err != nil {
			return err
		}
		if err := s.removeFromHashIndex(tx, "phash_index", fp.PHashes.PHash, imageID); err != nil {
			return err
		}
		if err := s.removeFromHashIndex(tx, "dhash_index", fp.PHashes.DHash, imageID); err != nil {
			return err
		}
		if err := s.removeFromHashIndex(tx, "whash_index", fp.PHashes.WHash, imageID); err != nil {
			return err
		}

		s.logger.Infof("Successfully deleted fingerprint: %s", imageID)
		return nil
	})
}

// removeFromHashIndex removes an image from a specific hash index
func (s *BoltStore) removeFromHashIndex(tx *bolt.Tx, bucketName string, hash uint64, imageID api.ImageID) error {
	if hash == 0 {
		return nil // Skip if hash wasn't computed
	}

	bucket := tx.Bucket([]byte(bucketName))
	key := fmt.Sprintf("%016x", hash)

	existingData := bucket.Get([]byte(key))
	if existingData == nil {
		return nil // No entries for this hash
	}

	var images []api.ImageID
	if err := json.Unmarshal(existingData, &images); err != nil {
		return fmt.Errorf("failed to unmarshal image list: %w", err)
	}

	// Filter out the image to remove
	var newImages []api.ImageID
	for _, id := range images {
		if id != imageID {
			newImages = append(newImages, id)
		}
	}

	if len(newImages) == 0 {
		// Remove the entire hash entry if no images left
		return bucket.Delete([]byte(key))
	}

	// Update with remaining images
	newData, err := json.Marshal(newImages)
	if err != nil {
		return fmt.Errorf("failed to marshal updated image list: %w", err)
	}

	return bucket.Put([]byte(key), newData)
}

// FindBySHA256 finds all images with a specific SHA256 hash
func (s *BoltStore) FindBySHA256(hash string) ([]api.ImageFingerprint, error) {
	var fingerprints []api.ImageFingerprint

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("sha256_index"))
		imageIDData := bucket.Get([]byte(hash))
		if imageIDData == nil {
			return nil // No images found, return empty slice
		}

		var imageIDs []api.ImageID
		if err := json.Unmarshal(imageIDData, &imageIDs); err != nil {
			// Handle single image ID (backward compatibility)
			imageIDs = []api.ImageID{api.ImageID(imageIDData)}
		}

		// Retrieve full fingerprints for each image ID
		fpBucket := tx.Bucket([]byte("fingerprints"))
		for _, imageID := range imageIDs {
			data := fpBucket.Get([]byte(imageID))
			if data != nil {
				var fp api.ImageFingerprint
				if err := json.Unmarshal(data, &fp); err != nil {
					s.logger.Warnf("Failed to unmarshal fingerprint %s: %v", imageID, err)
					continue
				}
				fingerprints = append(fingerprints, fp)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find by SHA256: %w", err)
	}

	return fingerprints, nil
}

// FindSimilarHashes finds images with similar perceptual hashes within maximum distance
func (s *BoltStore) FindSimilarHashes(targetHash uint64, maxDistance int, hashType string) ([]api.ImageFingerprint, error) {
	var similarFingerprints []api.ImageFingerprint

	err := s.db.View(func(tx *bolt.Tx) error {
		bucketName := hashType + "_index"
		bucket := tx.Bucket([]byte(bucketName))

		// Iterate through all hashes in the index
		return bucket.ForEach(func(k, v []byte) error {
			var storedHash uint64
			if _, err := fmt.Sscanf(string(k), "%x", &storedHash); err != nil {
				s.logger.Warnf("Invalid hash key format: %s", k)
				return nil
			}

			// Calculate Hamming distance between hashes
			distance := hammingDistance(targetHash, storedHash)
			if distance <= maxDistance {
				var imageIDs []api.ImageID
				if err := json.Unmarshal(v, &imageIDs); err != nil {
					s.logger.Warnf("Failed to unmarshal image IDs for hash %s: %v", k, err)
					return nil
				}

				// Retrieve fingerprints for similar images
				fpBucket := tx.Bucket([]byte("fingerprints"))
				for _, imageID := range imageIDs {
					data := fpBucket.Get([]byte(imageID))
					if data != nil {
						var fp api.ImageFingerprint
						if err := json.Unmarshal(data, &fp); err != nil {
							s.logger.Warnf("Failed to unmarshal fingerprint %s: %v", imageID, err)
							continue
						}
						similarFingerprints = append(similarFingerprints, fp)
					}
				}
			}

			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find similar hashes: %w", err)
	}

	return similarFingerprints, nil
}

// GetStats returns statistics about the image index
func (s *BoltStore) GetStats() (*Stats, error) {
	stats := &Stats{}

	err := s.db.View(func(tx *bolt.Tx) error {
		fingerprintsBucket := tx.Bucket([]byte("fingerprints"))
		if fingerprintsBucket == nil {
			// No fingerprints bucket yet → empty stats
			return nil
		}

		var totalSize int64
		var totalQuality float64
		var count int

		err := fingerprintsBucket.ForEach(func(k, v []byte) error {
			var fp api.ImageFingerprint
			if err := json.Unmarshal(v, &fp); err != nil {
				// Skip corrupted entries but don't stop the whole scan
				s.logger.Warnf("Failed to unmarshal fingerprint %s: %v", k, err)
				return nil
			}

			totalSize += fp.Metadata.SizeBytes
			totalQuality += fp.Quality.FinalScore
			count++

			return nil
		})
		if err != nil {
			return err
		}

		stats.TotalImages = int64(count)
		stats.TotalSizeBytes = totalSize
		if count > 0 {
			stats.AverageQuality = totalQuality / float64(count)
		}

		// NOTE: IndexSizeBytes is left as 0 for now.
		//  real on-disk size later
		// -  BoltStore
		// -  os.Stat(path).Size()

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to collect stats: %w", err)
	}

	return stats, nil
}

// Close safely closes the database connection
func (s *BoltStore) Close() error {
	s.logger.Info("Closing BoltDB index store")
	return s.db.Close()
}

// Compact performs database compaction (BoltDB handles this automatically)
func (s *BoltStore) Compact() error {
	// BoltDB doesn't require explicit compaction
	return nil
}

// hammingDistance calculates the Hamming distance between two 64-bit integers
func hammingDistance(a, b uint64) int {
	xor := a ^ b
	distance := 0
	for xor != 0 {
		distance++
		xor &= xor - 1
	}
	return distance
}
