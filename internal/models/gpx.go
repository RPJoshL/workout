package models

import "time"

// GpxFile represents a single file that contains GPS / workout data
// for exactly one workout
type GpxFile struct {
	DeviceData

	// Workout type provided within GPX file
	Type int `json:"type"`

	// Only for API request: name of the workout
	TypeName string `json:"typeName"`

	// Trackpoints with various values
	Points []GpxPoint `json:"points"`

	// Whether to use a higher sampling rate for downsampling (3s vs 6s)
	HigherSamplingRate bool `json:"useHighSamplingInterval"`
}

// DeviceData contains additional data that were tracked and calculated
// by the tracking device.
// This is useful for workout types that are not tracked by GPS
type DeviceData struct {
	UseDeviceData bool `json:"useDeviceData"`

	// Average speed in s/km
	SpeedAvg int `json:"speedAvg"`

	// Total distance in meters
	DistanceTotal int `json:"distanceTotal"`

	// Duration in seconds after which two points are considered as "paused"
	PauseDuration int `json:"pausedDuration"`
}

// GpxPoint is a waypoint of a GpxFile
type GpxPoint struct {
	// Latitude
	Lat float32 `json:"latitude"`

	// Longitude
	Lon float32 `json:"longitude"`

	// Horizontal accuracy in meters
	HorizontalAccuracy *float64 `json:"horizontalAccuracy"`

	// Elevation in full meters
	Elevation int `json:"elevation"`

	// Timestamp of this workout point
	Timestamp time.Time `json:"time"`

	// Heart rate
	HeartRate int `json:"heartrate"`

	// Total number of steps since the beginning of the workout
	Steps int `json:"steps"`

	// Speed in s/km
	Speed int `json:"speed"`

	// Total distance in meters since the beginning of the workout
	Distance int `json:"distance"`

	// Acceleration data for this point as a repeating array:
	//  - relative timestamp in ms for point
	//  - x,y,z stored as signed Int16, scaled in g: 2048 = 1 g (9.80665 m/s²)
	Acceleration []int16 `json:"acceleration"`
}

// EqualValues returns whether [pp] has the same values (position, heartrate)
// as the provided one
func (p GpxPoint) EqualValues(pp GpxPoint) bool {
	return (p.Lat == pp.Lat && p.Lon == pp.Lon && p.Elevation == pp.Elevation && p.HeartRate == pp.HeartRate && p.Distance == pp.Distance) ||
		(p.HeartRate > 20 && p.HeartRate == pp.HeartRate && p.Distance == pp.Distance)
}
