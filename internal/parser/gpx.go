package parser

import "math"

// EarthRadius is the radius of the earch in meters
const EarthRadius = 6371000

// RadianToDegree konvertiert Bogenmaß in Grad
func RadianToDegree(rad float64) float64 {
	return rad * 180 / math.Pi
}

// DegreeToRadian konvertiert Grad in Bogenmaß
func DegreeToRadian(deg float64) float64 {
	return deg * math.Pi / 180
}

// GetDestinationPosition berechnet die Zielposition basierend auf Startposition, Abstand und Winkel
func getDestinationPosition(lat, lon, distance, bearing float64) (destLat float64, destLon float64) {

	// Konvertiere Winkel in Bogenmaß
	bearing = DegreeToRadian(bearing)

	// Konvertiere Breitengrad und Längengrad in Bogenmaß
	latRad := DegreeToRadian(lat)
	lonRad := DegreeToRadian(lon)

	// Calculate the target latitude
	destLat = math.Asin(math.Sin(latRad)*math.Cos(distance/EarthRadius) +
		math.Cos(latRad)*math.Sin(distance/EarthRadius)*math.Cos(bearing))

	// Calculate the target length difference
	destLonDiff := math.Atan2(math.Sin(bearing)*math.Sin(distance/EarthRadius)*math.Cos(latRad),
		math.Cos(distance/EarthRadius)-math.Sin(latRad)*math.Sin(destLat))

	// Calculate the target degree of length
	destLon = lonRad + destLonDiff

	// Convert target latitude and target longitude to degrees
	destLat = RadianToDegree(destLat)
	destLon = RadianToDegree(destLon)

	return
}

// AddMetersToPosition calculates a new position based on a starting position and the distance in meters
// The 'horizontal' parameter specifies whether the distance is added horizontally (true) or vertically (false).
func AddMetersToPosition(startLat, startLon, meters float64, horizontal bool) (destLat float64, destLon float64) {
	var bearing float64
	if horizontal {
		bearing = 90.0
	} else {
		bearing = 0.0
	}
	return getDestinationPosition(startLat, startLon, meters, bearing)
}
