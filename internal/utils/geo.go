package utils

import (
	"fmt"
	"math"
)

// Haversine calculates the distance between two points on Earth in kilometers
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// GetLocationClusterKey generates a cluster key for location-based caching
func GetLocationClusterKey(lat, lon, clusterDegrees float64) string {
	clusterLat := math.Floor(lat/clusterDegrees) * clusterDegrees
	clusterLon := math.Floor(lon/clusterDegrees) * clusterDegrees
	return fmt.Sprintf("%.2f_%.2f", clusterLat, clusterLon)
}
