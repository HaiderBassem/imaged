package similarity

import (
	"fmt"
	"hash/fnv"
	"math/rand"
)

// LSH implements Locality Sensitive Hashing for efficient similarity search
type LSH struct {
	numTables     int
	numHashes     int
	hashTables    []map[uint32][]string
	hashFunctions []func([]float64) uint32
}

// NewLSH creates a new LSH index
func NewLSH(numTables, numHashes int) *LSH {
	lsh := &LSH{
		numTables:  numTables,
		numHashes:  numHashes,
		hashTables: make([]map[uint32][]string, numTables),
	}

	// Initialize hash tables
	for i := 0; i < numTables; i++ {
		lsh.hashTables[i] = make(map[uint32][]string)
	}

	// Generate random hash functions
	lsh.generateHashFunctions()

	return lsh
}

// generateHashFunctions creates random projection hash functions
func (l *LSH) generateHashFunctions() {
	l.hashFunctions = make([]func([]float64) uint32, l.numTables*l.numHashes)

	for i := 0; i < l.numTables*l.numHashes; i++ {
		// Create a random projection vector
		l.hashFunctions[i] = l.createRandomProjection()
	}
}

// createRandomProjection creates a random projection hash function
func (l *LSH) createRandomProjection() func([]float64) uint32 {
	// Generate random weights for the projection
	return func(vector []float64) uint32 {
		var hash uint32
		fmt.Print(hash)
		hasher := fnv.New32a()

		for i, value := range vector {
			// Simple random projection (in practice, use better random values)
			weight := rand.Float64()*2 - 1 // Random between -1 and 1
			projected := value * weight

			// Threshold to get binary value
			if projected > 0 {
				hasher.Write([]byte{byte(i), 1})
			} else {
				hasher.Write([]byte{byte(i), 0})
			}
		}

		return hasher.Sum32()
	}
}

// IndexVector indexes a vector with its ID
func (l *LSH) IndexVector(vector []float64, id string) {
	for table := 0; table < l.numTables; table++ {
		// Compute combined hash for this table
		tableHash := l.computeTableHash(vector, table)

		// Add to hash table
		l.hashTables[table][tableHash] = append(l.hashTables[table][tableHash], id)
	}
}

// computeTableHash computes the combined hash for a table
func (l *LSH) computeTableHash(vector []float64, table int) uint32 {
	hasher := fnv.New32a()

	for hash := 0; hash < l.numHashes; hash++ {
		funcIndex := table*l.numHashes + hash
		hashValue := l.hashFunctions[funcIndex](vector)
		hasher.Write([]byte{
			byte(hashValue >> 24),
			byte(hashValue >> 16),
			byte(hashValue >> 8),
			byte(hashValue),
		})
	}

	return hasher.Sum32()
}

// Query finds similar vectors using LSH
func (l *LSH) Query(vector []float64, maxResults int) []string {
	candidates := make(map[string]bool)

	for table := 0; table < l.numTables; table++ {
		tableHash := l.computeTableHash(vector, table)

		if bucket, exists := l.hashTables[table][tableHash]; exists {
			for _, id := range bucket {
				candidates[id] = true
			}
		}
	}

	// Convert to slice
	result := make([]string, 0, len(candidates))
	for id := range candidates {
		result = append(result, id)
		if len(result) >= maxResults {
			break
		}
	}

	return result
}

// GetBucketStats returns statistics about hash table buckets
func (l *LSH) GetBucketStats() map[string]interface{} {
	stats := make(map[string]interface{})

	var totalBuckets int
	var totalItems int
	var maxBucketSize int

	for _, table := range l.hashTables {
		totalBuckets += len(table)
		for _, items := range table {
			totalItems += len(items)
			if len(items) > maxBucketSize {
				maxBucketSize = len(items)
			}
		}
	}

	stats["num_tables"] = l.numTables
	stats["total_buckets"] = totalBuckets
	stats["total_items"] = totalItems
	stats["max_bucket_size"] = maxBucketSize

	if totalBuckets > 0 {
		stats["avg_bucket_size"] = float64(totalItems) / float64(totalBuckets)
	} else {
		stats["avg_bucket_size"] = 0.0
	}

	return stats
}
