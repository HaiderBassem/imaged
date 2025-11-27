package similarity

import (
	"fmt"

	"github.com/HaiderBassem/imaged/pkg/api"
)

// Grouper handles grouping of similar images
type Grouper struct {
	comparator *Comparator
}

// NewGrouper creates a new image grouper
func NewGrouper(comparator *Comparator) *Grouper {
	return &Grouper{
		comparator: comparator,
	}
}

// GroupBySimilarity groups images by similarity threshold
func (g *Grouper) GroupBySimilarity(fingerprints []api.ImageFingerprint, threshold float64) []api.DuplicateGroup {
	var groups []api.DuplicateGroup
	assigned := make(map[api.ImageID]bool)
	groupID := 0

	for i, fp1 := range fingerprints {
		if assigned[fp1.ID] {
			continue
		}

		group := api.DuplicateGroup{
			GroupID:   generateGroupID(groupID),
			MainImage: fp1.ID,
			Reason:    "similarity",
		}
		groupID++

		assigned[fp1.ID] = true

		for j := i + 1; j < len(fingerprints); j++ {
			fp2 := fingerprints[j]
			if assigned[fp2.ID] {
				continue
			}

			similarity, err := g.comparator.CompareFingerprints(fp1, fp2)
			if err != nil {
				continue
			}

			if similarity >= threshold {
				group.DuplicateIDs = append(group.DuplicateIDs, fp2.ID)
				assigned[fp2.ID] = true
				group.Confidence = similarity
			}
		}

		if len(group.DuplicateIDs) > 0 {
			groups = append(groups, group)
		}
	}

	return groups
}

// ClusterByContent clusters images based on content similarity
func (g *Grouper) ClusterByContent(fingerprints []api.ImageFingerprint, threshold float64) []api.Cluster {
	var clusters []api.Cluster
	assigned := make(map[api.ImageID]bool)
	clusterID := 0

	for i, fp1 := range fingerprints {
		if assigned[fp1.ID] {
			continue
		}

		cluster := api.Cluster{
			ClusterID: generateClusterID(clusterID),
			Images:    []api.ImageID{fp1.ID},
		}
		clusterID++

		assigned[fp1.ID] = true

		for j := i + 1; j < len(fingerprints); j++ {
			fp2 := fingerprints[j]
			if assigned[fp2.ID] {
				continue
			}

			similarity, err := g.comparator.CompareFingerprints(fp1, fp2)
			if err != nil {
				continue
			}

			if similarity >= threshold {
				cluster.Images = append(cluster.Images, fp2.ID)
				assigned[fp2.ID] = true
			}
		}

		if len(cluster.Images) > 0 {
			clusters = append(clusters, cluster)
		}
	}

	return clusters
}

// generateGroupID creates a unique group identifier
func generateGroupID(id int) string {
	return fmt.Sprintf("group_%d", id)
}

// generateClusterID creates a unique cluster identifier
func generateClusterID(id int) string {
	return fmt.Sprintf("cluster_%d", id)
}
