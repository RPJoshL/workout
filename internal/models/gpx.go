package models

import "time"

// GpxFile represents a single file that contains GPS / workout data
// for exactly one workout
type GpxFile struct {

	// Workout type provided within GPX file
	Type int

	// Trackpoints with various values
	Points []GpxPoint
}

// GpxPoint is a waypoint of a GpxFile
type GpxPoint struct {

	// Latitude
	Lat float32

	// Longitude
	Lon float32

	// Elevation in full meters
	Elevation int

	// Timestamp of this workout point
	Timestamp time.Time

	// Heart rate
	HeartRate int
}
