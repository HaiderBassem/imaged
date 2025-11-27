package similarity

import (
	"github.com/HaiderBassem/imaged/pkg/api"
)

// Clusterer performs advanced image clustering
type Clusterer struct {
	comparator *Comparator
}

// NewClusterer creates a new image clusterer
func NewClusterer(comparator *Comparator) *Clusterer {
	return &Clusterer{
		comparator: comparator,
	}
}

// ClusterBySimilarity performs hierarchical clustering
func (c *Clusterer) ClusterBySimilarity(fingerprints []api.ImageFingerprint, threshold float64) []api.Cluster {
	if len(fingerprints) == 0 {
		return []api.Cluster{}
	}

	// Start with each image in its own cluster
	clusters := c.initializeClusters(fingerprints)

	// Perform hierarchical clustering
	for {
		if len(clusters) <= 1 {
			break
		}

		// Find closest clusters
		closestI, closestJ, maxSimilarity := c.findClosestClusters(clusters)

		if maxSimilarity < threshold {
			break // No more clusters to merge
		}

		// Merge clusters
		clusters = c.mergeClusters(clusters, closestI, closestJ)
	}

	return clusters
}

// initializeClusters creates initial single-element clusters
func (c *Clusterer) initializeClusters(fingerprints []api.ImageFingerprint) []api.Cluster {
	clusters := make([]api.Cluster, len(fingerprints))
	for i, fp := range fingerprints {
		clusters[i] = api.Cluster{
			ClusterID: generateClusterID(i),
			Images:    []api.ImageID{fp.ID},
		}
	}
	return clusters
}

// findClosestClusters finds the two most similar clusters
func (c *Clusterer) findClosestClusters(clusters []api.Cluster) (int, int, float64) {
	var maxSimilarity float64
	var closestI, closestJ int

	for i := 0; i < len(clusters); i++ {
		for j := i + 1; j < len(clusters); j++ {
			similarity := c.clusterSimilarity(clusters[i], clusters[j])
			if similarity > maxSimilarity {
				maxSimilarity = similarity
				closestI = i
				closestJ = j
			}
		}
	}

	return closestI, closestJ, maxSimilarity
}

// clusterSimilarity calculates similarity between two clusters
func (c *Clusterer) clusterSimilarity(cluster1, cluster2 api.Cluster) float64 {
	var totalSimilarity float64
	var comparisons int

	// Use average linkage (average similarity between all pairs)
	for _, img1 := range cluster1.Images {
		for _, img2 := range cluster2.Images {
			// In a real implementation, you would compare the actual fingerprints
			// For now, we'll use a simplified approach
			fp1 := c.findFingerprint(img1)
			fp2 := c.findFingerprint(img2)

			if fp1 != nil && fp2 != nil {
				similarity, err := c.comparator.CompareFingerprints(*fp1, *fp2)
				if err == nil {
					totalSimilarity += similarity
					comparisons++
				}
			}
		}
	}

	if comparisons == 0 {
		return 0.0
	}

	return totalSimilarity / float64(comparisons)
}

// mergeClusters merges two clusters
func (c *Clusterer) mergeClusters(clusters []api.Cluster, i, j int) []api.Cluster {
	// Merge cluster j into cluster i
	clusters[i].Images = append(clusters[i].Images, clusters[j].Images...)

	// Remove cluster j
	return append(clusters[:j], clusters[j+1:]...)
}

// DBSCANClustering performs density-based clustering
func (c *Clusterer) DBSCANClustering(fingerprints []api.ImageFingerprint, eps float64, minPts int) []api.Cluster {
	visited := make(map[api.ImageID]bool)
	clusters := []api.Cluster{}
	clusterID := 0

	for _, fp := range fingerprints {
		if visited[fp.ID] {
			continue
		}
		visited[fp.ID] = true

		neighbors := c.rangeQuery(fingerprints, fp, eps)

		if len(neighbors) < minPts {
			// Noise point (will be handled later)
			continue
		}

		// Expand cluster
		cluster := api.Cluster{
			ClusterID: generateClusterID(clusterID),
			Images:    []api.ImageID{fp.ID},
		}
		clusterID++

		neighbors = c.expandCluster(fingerprints, neighbors, cluster, eps, minPts, visited)
		clusters = append(clusters, cluster)
	}

	return clusters
}

// rangeQuery finds all points within epsilon distance
func (c *Clusterer) rangeQuery(fingerprints []api.ImageFingerprint, point api.ImageFingerprint, eps float64) []api.ImageFingerprint {
	var neighbors []api.ImageFingerprint

	for _, fp := range fingerprints {
		similarity, err := c.comparator.CompareFingerprints(point, fp)
		if err == nil && similarity >= eps {
			neighbors = append(neighbors, fp)
		}
	}

	return neighbors
}

// expandCluster expands a cluster based on density
func (c *Clusterer) expandCluster(fingerprints, seeds []api.ImageFingerprint, cluster api.Cluster, eps float64, minPts int, visited map[api.ImageID]bool) []api.ImageFingerprint {
	for i := 0; i < len(seeds); i++ {
		point := seeds[i]

		if !visited[point.ID] {
			visited[point.ID] = true
			cluster.Images = append(cluster.Images, point.ID)

			pointNeighbors := c.rangeQuery(fingerprints, point, eps)
			if len(pointNeighbors) >= minPts {
				seeds = append(seeds, pointNeighbors...)
			}
		}
	}

	return seeds
}

// KMeansClustering performs K-means clustering (simplified)
func (c *Clusterer) KMeansClustering(fingerprints []api.ImageFingerprint, k int, maxIterations int) []api.Cluster {
	if len(fingerprints) == 0 || k <= 0 {
		return []api.Cluster{}
	}

	k = min(k, len(fingerprints))

	// Initialize centroids randomly
	centroids := c.initializeCentroids(fingerprints, k)
	clusters := make([]api.Cluster, k)

	for iter := 0; iter < maxIterations; iter++ {
		// Reset clusters
		for i := range clusters {
			clusters[i] = api.Cluster{
				ClusterID: generateClusterID(i),
				Images:    []api.ImageID{},
			}
		}

		// Assign points to nearest centroid
		for _, fp := range fingerprints {
			nearestCentroid := c.findNearestCentroid(fp, centroids)
			clusters[nearestCentroid].Images = append(clusters[nearestCentroid].Images, fp.ID)
		}

		// Update centroids
		newCentroids := c.updateCentroids(clusters, fingerprints)

		// Check for convergence
		if c.centroidsConverged(centroids, newCentroids) {
			break
		}

		centroids = newCentroids
	}

	return clusters
}

// initializeCentroids initializes K centroids randomly
func (c *Clusterer) initializeCentroids(fingerprints []api.ImageFingerprint, k int) []api.ImageFingerprint {
	centroids := make([]api.ImageFingerprint, k)

	// Simple random selection (in practice, use better initialization like k-means++)
	for i := 0; i < k; i++ {
		centroids[i] = fingerprints[i%len(fingerprints)]
	}

	return centroids
}

// findNearestCentroid finds the nearest centroid for a point
func (c *Clusterer) findNearestCentroid(point api.ImageFingerprint, centroids []api.ImageFingerprint) int {
	var minDistance float64 = -1
	nearest := 0

	for i, centroid := range centroids {
		similarity, err := c.comparator.CompareFingerprints(point, centroid)
		if err != nil {
			continue
		}

		distance := 1 - similarity
		if minDistance < 0 || distance < minDistance {
			minDistance = distance
			nearest = i
		}
	}

	return nearest
}

// updateCentroids updates cluster centroids
func (c *Clusterer) updateCentroids(clusters []api.Cluster, fingerprints []api.ImageFingerprint) []api.ImageFingerprint {
	centroids := make([]api.ImageFingerprint, len(clusters))

	for i, cluster := range clusters {
		if len(cluster.Images) == 0 {
			// Keep current centroid if cluster is empty
			continue
		}

		// Find the most central point in the cluster
		centroids[i] = c.findMedoid(cluster, fingerprints)
	}

	return centroids
}

// findMedoid finds the most central point in a cluster
func (c *Clusterer) findMedoid(cluster api.Cluster, fingerprints []api.ImageFingerprint) api.ImageFingerprint {
	var bestFingerprint api.ImageFingerprint
	var minTotalDistance float64 = -1

	for _, imgID := range cluster.Images {
		fp := c.findFingerprint(imgID)
		if fp == nil {
			continue
		}

		totalDistance := 0.0
		validComparisons := 0

		for _, otherID := range cluster.Images {
			if otherID == imgID {
				continue
			}

			otherFP := c.findFingerprint(otherID)
			if otherFP != nil {
				similarity, err := c.comparator.CompareFingerprints(*fp, *otherFP)
				if err == nil {
					totalDistance += 1 - similarity
					validComparisons++
				}
			}
		}

		if validComparisons > 0 {
			avgDistance := totalDistance / float64(validComparisons)
			if minTotalDistance < 0 || avgDistance < minTotalDistance {
				minTotalDistance = avgDistance
				bestFingerprint = *fp
			}
		}
	}

	return bestFingerprint
}

// centroidsConverged checks if centroids have converged
func (c *Clusterer) centroidsConverged(oldCentroids, newCentroids []api.ImageFingerprint) bool {
	if len(oldCentroids) != len(newCentroids) {
		return false
	}

	for i := range oldCentroids {
		similarity, err := c.comparator.CompareFingerprints(oldCentroids[i], newCentroids[i])
		if err != nil || similarity < 0.99 { // 99% similarity threshold
			return false
		}
	}

	return true
}

// findFingerprint helper function (would be connected to index in real implementation)
func (c *Clusterer) findFingerprint(imageID api.ImageID) *api.ImageFingerprint {
	// This would query the index in a real implementation
	return nil
}

// Utility functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
