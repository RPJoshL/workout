package models

import "time"

// GpxFile represents a single file that contains GPS / workout data
// for exactly one workout
type GpxFile struct {

	// Workout type provided within GPX file
	Type int `json:"type"`

	// Only for API request: name of the workout
	TypeName string `json:"typeName"`

	// Trackpoints with various values
	Points []GpxPoint `json:"points"`
}

// GpxPoint is a waypoint of a GpxFile
type GpxPoint struct {

	// Latitude
	Lat float32 `json:"latitude"`

	// Longitude
	Lon float32 `json:"longitude"`

	// Elevation in full meters
	Elevation int `json:"elevation"`

	// Timestamp of this workout point
	Timestamp time.Time `json:"time"`

	// Heart rate
	HeartRate int `json:"heartrate"`

	// Total number of steps since the beginning of the workout
	Steps int `json:"steps"`
}

// EqualValues returns whether [pp] has the same values (position, heartrate)
// as the provided one
func (p GpxPoint) EqualValues(pp GpxPoint) bool {
	return (p.Lat == pp.Lat && p.Lon == pp.Lon && p.Elevation == pp.Elevation && p.HeartRate == pp.HeartRate) ||
		(p.HeartRate > 20 && p.HeartRate == pp.HeartRate)
}
