package utils

import (
	"math"
)

const EarthRadiusMeters = 6371000.0 // Mean radius of the Earth in meters

// HaversineDistance calculates the great-circle distance in meters between two coordinates.
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusMeters * c
}

func degreesToRadians(deg float64) float64 {
	return deg * (math.Pi / 180.0)
}
