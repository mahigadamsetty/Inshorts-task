package utils

import (
	"strings"
)

// CalculateTextMatchScore calculates a simple text matching score
// Returns a score between 0 and 1 based on keyword presence
func CalculateTextMatchScore(query, title, description string) float64 {
	queryLower := strings.ToLower(query)
	titleLower := strings.ToLower(title)
	descLower := strings.ToLower(description)
	
	// Split query into keywords
	keywords := strings.Fields(queryLower)
	if len(keywords) == 0 {
		return 0.0
	}
	
	matchCount := 0
	totalWeight := 0.0
	
	for _, keyword := range keywords {
		if len(keyword) < 2 {
			continue
		}
		
		// Title matches count more (weight 2)
		if strings.Contains(titleLower, keyword) {
			matchCount++
			totalWeight += 2.0
		}
		
		// Description matches count less (weight 1)
		if strings.Contains(descLower, keyword) {
			matchCount++
			totalWeight += 1.0
		}
	}
	
	// Normalize score to 0-1 range
	maxPossibleWeight := float64(len(keywords)) * 3.0 // max weight if all keywords match in both
	if maxPossibleWeight == 0 {
		return 0.0
	}
	
	return totalWeight / maxPossibleWeight
}
