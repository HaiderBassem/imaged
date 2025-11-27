package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HaiderBassem/imaged/pkg/api"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// SQLiteStore implements the Store interface using SQLite
type SQLiteStore struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewSQLiteStore creates a new SQLite-based index store
func NewSQLiteStore(dbPath string) (Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	store := &SQLiteStore{
		db:     db,
		logger: logger,
	}

	// Initialize database schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// GetStats returns statistics about the SQLite index
func (s *SQLiteStore) GetStats() (*Stats, error) {
	var stats Stats

	// Total images and size
	row := s.db.QueryRow(`
		SELECT 
			COUNT(*),
			IFNULL(SUM(json_extract(metadata, '$.SizeBytes')), 0)
		FROM fingerprints
	`)

	var totalImages int64
	var totalSize int64

	if err := row.Scan(&totalImages, &totalSize); err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}

	// Average quality
	row = s.db.QueryRow(`
		SELECT 
			IFNULL(AVG(json_extract(quality, '$.FinalScore')), 0)
		FROM fingerprints
	`)

	var avgQuality float64
	if err := row.Scan(&avgQuality); err != nil {
		return nil, fmt.Errorf("failed to query average quality: %w", err)
	}

	stats.TotalImages = totalImages
	stats.TotalSizeBytes = totalSize
	stats.AverageQuality = avgQuality

	// SQLite total database size (approximate)
	var pageCount int64
	var pageSize int64

	if err := s.db.QueryRow(`PRAGMA page_count`).Scan(&pageCount); err != nil {
		return nil, fmt.Errorf("failed to get page_count: %w", err)
	}

	if err := s.db.QueryRow(`PRAGMA page_size`).Scan(&pageSize); err != nil {
		return nil, fmt.Errorf("failed to get page_size: %w", err)
	}

	stats.IndexSizeBytes = pageCount * pageSize

	return &stats, nil
}

// initSchema creates the necessary database tables
func (s *SQLiteStore) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS fingerprints (
            id TEXT PRIMARY KEY,
            metadata TEXT NOT NULL,
            phashes TEXT NOT NULL,
            quality TEXT NOT NULL,
            color_hist TEXT,
            feature_vec TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS sha256_index (
            sha256 TEXT PRIMARY KEY,
            image_id TEXT NOT NULL,
            FOREIGN KEY (image_id) REFERENCES fingerprints (id)
        )`,
		`CREATE TABLE IF NOT EXISTS perceptual_index (
            hash_type TEXT,
            hash_value INTEGER,
            image_id TEXT,
            PRIMARY KEY (hash_type, hash_value, image_id),
            FOREIGN KEY (image_id) REFERENCES fingerprints (id)
        )`,
		`CREATE TABLE IF NOT EXISTS path_index (
            path TEXT PRIMARY KEY,
            image_id TEXT NOT NULL,
            FOREIGN KEY (image_id) REFERENCES fingerprints (id)
        )`,
		`CREATE INDEX IF NOT EXISTS idx_ahash ON perceptual_index(hash_type, hash_value)`,
		`CREATE INDEX IF NOT EXISTS idx_phash ON perceptual_index(hash_type, hash_value)`,
		`CREATE INDEX IF NOT EXISTS idx_dhash ON perceptual_index(hash_type, hash_value)`,
		`CREATE INDEX IF NOT EXISTS idx_whash ON perceptual_index(hash_type, hash_value)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}
	return nil
}

// SaveFingerprint stores an image fingerprint
func (s *SQLiteStore) SaveFingerprint(fp api.ImageFingerprint) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	metadataJSON, _ := json.Marshal(fp.Metadata)
	phashesJSON, _ := json.Marshal(fp.PHashes)
	qualityJSON, _ := json.Marshal(fp.Quality)

	var colorHistJSON []byte
	if fp.ColorHist != nil {
		colorHistJSON, _ = json.Marshal(fp.ColorHist)
	}

	var featureVecJSON []byte
	if fp.FeatureVec != nil {
		featureVecJSON, _ = json.Marshal(fp.FeatureVec)
	}

	_, err = tx.Exec(`
        INSERT OR REPLACE INTO fingerprints 
        (id, metadata, phashes, quality, color_hist, feature_vec, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, string(fp.ID), string(metadataJSON), string(phashesJSON),
		string(qualityJSON), colorHistJSON, featureVecJSON, fp.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert fingerprint: %w", err)
	}

	_, err = tx.Exec(`INSERT OR REPLACE INTO sha256_index (sha256, image_id) VALUES (?, ?)`,
		fp.Metadata.SHA256, string(fp.ID))
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT OR REPLACE INTO path_index (path, image_id) VALUES (?, ?)`,
		fp.Metadata.Path, string(fp.ID))
	if err != nil {
		return err
	}

	if err := s.updatePerceptualIndex(tx, fp); err != nil {
		return err
	}

	return tx.Commit()
}

// updatePerceptualIndex updates perceptual hash indices
func (s *SQLiteStore) updatePerceptualIndex(tx *sql.Tx, fp api.ImageFingerprint) error {
	_, err := tx.Exec("DELETE FROM perceptual_index WHERE image_id = ?", string(fp.ID))
	if err != nil {
		return err
	}

	hashes := map[string]uint64{
		"ahash": fp.PHashes.AHash,
		"phash": fp.PHashes.PHash,
		"dhash": fp.PHashes.DHash,
		"whash": fp.PHashes.WHash,
	}

	for hashType, hashValue := range hashes {
		if hashValue == 0 {
			continue
		}
		_, err := tx.Exec(`
            INSERT INTO perceptual_index (hash_type, hash_value, image_id)
            VALUES (?, ?, ?)
        `, hashType, hashValue, string(fp.ID))
		if err != nil {
			return err
		}
	}

	return nil
}

// GetFingerprint retrieves a fingerprint by ID
func (s *SQLiteStore) GetFingerprint(imageID api.ImageID) (*api.ImageFingerprint, error) {
	var fp api.ImageFingerprint
	var metadataJSON, phashesJSON, qualityJSON, colorHistJSON, featureVecJSON string
	var createdAt time.Time

	err := s.db.QueryRow(`
        SELECT id, metadata, phashes, quality, color_hist, feature_vec, created_at
        FROM fingerprints WHERE id = ?
    `, string(imageID)).Scan(
		&fp.ID, &metadataJSON, &phashesJSON, &qualityJSON,
		&colorHistJSON, &featureVecJSON, &createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, api.ErrImageNotFound
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(metadataJSON), &fp.Metadata)
	json.Unmarshal([]byte(phashesJSON), &fp.PHashes)
	json.Unmarshal([]byte(qualityJSON), &fp.Quality)

	if colorHistJSON != "" {
		json.Unmarshal([]byte(colorHistJSON), &fp.ColorHist)
	}
	if featureVecJSON != "" {
		json.Unmarshal([]byte(featureVecJSON), &fp.FeatureVec)
	}

	fp.CreatedAt = createdAt
	return &fp, nil
}

// GetAllFingerprints retrieves all fingerprints
func (s *SQLiteStore) GetAllFingerprints() ([]api.ImageFingerprint, error) {
	rows, err := s.db.Query(`SELECT id, metadata, phashes, quality, color_hist, feature_vec, created_at FROM fingerprints`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fingerprints []api.ImageFingerprint

	for rows.Next() {
		var fp api.ImageFingerprint
		var metadataJSON, phashesJSON, qualityJSON, colorHistJSON, featureVecJSON string
		var createdAt time.Time

		if err := rows.Scan(&fp.ID, &metadataJSON, &phashesJSON, &qualityJSON, &colorHistJSON, &featureVecJSON, &createdAt); err != nil {
			continue
		}

		json.Unmarshal([]byte(metadataJSON), &fp.Metadata)
		json.Unmarshal([]byte(phashesJSON), &fp.PHashes)
		json.Unmarshal([]byte(qualityJSON), &fp.Quality)

		fp.CreatedAt = createdAt
		fingerprints = append(fingerprints, fp)
	}

	return fingerprints, nil
}

// FindBySHA256 finds fingerprints by hash
func (s *SQLiteStore) FindBySHA256(hash string) ([]api.ImageFingerprint, error) {
	rows, err := s.db.Query(`
        SELECT f.id, f.metadata, f.phashes, f.quality, f.color_hist, f.feature_vec, f.created_at
        FROM fingerprints f
        JOIN sha256_index s ON f.id = s.image_id
        WHERE s.sha256 = ?
    `, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanFingerprints(rows)
}

// FindSimilarHashes finds similar perceptual hashes
func (s *SQLiteStore) FindSimilarHashes(targetHash uint64, maxDistance int, hashType string) ([]api.ImageFingerprint, error) {
	all, err := s.GetAllFingerprints()
	if err != nil {
		return nil, err
	}

	var similar []api.ImageFingerprint
	for _, fp := range all {
		var hv uint64
		switch hashType {
		case "ahash":
			hv = fp.PHashes.AHash
		case "phash":
			hv = fp.PHashes.PHash
		case "dhash":
			hv = fp.PHashes.DHash
		case "whash":
			hv = fp.PHashes.WHash
		}
		if hv == 0 {
			continue
		}
		if hammingDistance(targetHash, hv) <= maxDistance {
			similar = append(similar, fp)
		}
	}

	return similar, nil
}

// scanFingerprints helper
func (s *SQLiteStore) scanFingerprints(rows *sql.Rows) ([]api.ImageFingerprint, error) {
	var fingerprints []api.ImageFingerprint

	for rows.Next() {
		var fp api.ImageFingerprint
		var metadataJSON, phashesJSON, qualityJSON, colorHistJSON, featureVecJSON string
		var createdAt time.Time

		if err := rows.Scan(&fp.ID, &metadataJSON, &phashesJSON, &qualityJSON, &colorHistJSON, &featureVecJSON, &createdAt); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(metadataJSON), &fp.Metadata)
		json.Unmarshal([]byte(phashesJSON), &fp.PHashes)
		json.Unmarshal([]byte(qualityJSON), &fp.Quality)

		fp.CreatedAt = createdAt
		fingerprints = append(fingerprints, fp)
	}

	return fingerprints, nil
}

// Close closes database
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// Compact vacuum database
func (s *SQLiteStore) Compact() error {
	_, err := s.db.Exec("VACUUM")
	return err
}

// // hammingDistance helper
// func hammingDistance(a, b uint64) int {
// 	var dist int
// 	x := a ^ b
// 	for x != 0 {
// 		dist++
// 		x &= x - 1
// 	}
// 	return dist
// }

// DeleteFingerprint removes a fingerprint from SQLite
func (s *SQLiteStore) DeleteFingerprint(imageID api.ImageID) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get fingerprint first (to clean indices)
	fp, err := s.GetFingerprint(imageID)
	if err != nil {
		return err
	}

	// Delete from fingerprints table
	_, err = tx.Exec(`DELETE FROM fingerprints WHERE id = ?`, string(imageID))
	if err != nil {
		return fmt.Errorf("failed to delete fingerprint: %w", err)
	}

	// Delete from sha256 index
	_, err = tx.Exec(`DELETE FROM sha256_index WHERE sha256 = ?`, fp.Metadata.SHA256)
	if err != nil {
		return fmt.Errorf("failed to delete sha256 index: %w", err)
	}

	// Delete from path index
	_, err = tx.Exec(`DELETE FROM path_index WHERE path = ?`, fp.Metadata.Path)
	if err != nil {
		return fmt.Errorf("failed to delete path index: %w", err)
	}

	// Delete perceptual hashes
	_, err = tx.Exec(`DELETE FROM perceptual_index WHERE image_id = ?`, string(imageID))
	if err != nil {
		return fmt.Errorf("failed to delete perceptual index: %w", err)
	}

	return tx.Commit()
}
