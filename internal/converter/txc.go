package converter

import (
	"bytes"
	"encoding/xml"
	"math"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

var (
	ErrTxcError = errors.BadRequest("#workout.tcxError")
)

// Tcx is the root struct of a TCX file
type Tcx struct {
	XMLName      xml.Name      `xml:"TrainingCenterDatabase"`
	XMLNs        string        `xml:"xmlns,attr"`
	XMLNsXsi     string        `xml:"xsi,attr,omitempty"`
	XMLNsXsd     string        `xml:"xsd,attr,omitempty"`
	XMLSchemaLoc string        `xml:"schemaLocation,attr,omitempty"`
	Activities   []TxcActivity `xml:"Activities>Activity"`
}

type TxcActivity struct {
	Sport string    `xml:"Sport,attr"`
	ID    time.Time `xml:"Id"`
	Laps  []TxcLap  `xml:"Lap"`
}

type TxcLap struct {
	StartTime                  time.Time       `xml:"StartTime,attr"`
	TotalTimeInSeconds         float64         `xml:"TotalTimeSeconds"`
	DistanceInMeters           float64         `xml:"DistanceMeters"`
	MaximumSpeedInMetersPerSec float64         `xml:"MaximumSpeed"`
	Calories                   float64         `xml:"Calories"`
	Intensity                  string          `xml:"Intensity"`
	TriggerMethod              string          `xml:"TriggerMethod"`
	Track                      []TcxTrackpoint `xml:"Track>Trackpoint"`
}

type TcxTrackpoint struct {
	Time               time.Time `xml:"Time"`
	LatitudeInDegrees  float64   `xml:"Position>LatitudeDegrees"`
	LongitudeInDegrees float64   `xml:"Position>LongitudeDegrees"`
	AltitudeInMeters   float64   `xml:"AltitudeMeters"`
	HeartRateInBpm     int       `xml:"HeartRateBpm>Value"`
}

// ParseTcx parses the file content of a TCX file and return the
// transformed GPX struct
func ParseTcx(content []byte) (*models.GpxFile, errors.Error) {

	// Parse the file
	tcx := Tcx{}
	decoder := xml.NewDecoder(bytes.NewBuffer(content))
	if err := decoder.Decode(&tcx); err != nil {
		logger.Error("Unable to decode TCX file: %s", err)
		return nil, ErrTxcError
	}

	// Transform to generic GpxFile
	rtc := &models.GpxFile{}

	// Try to get a workout name
	if len(tcx.Activities) > 0 {
		rtc.Type = models.GetWorkoutTypeByName(tcx.Activities[0].Sport)
	}

	// Collect all Activities and tracks into a single GPX struct
	for _, activity := range tcx.Activities {
		for _, lap := range activity.Laps {
			for _, track := range lap.Track {
				// Basic TCX data
				p := models.GpxPoint{
					Lat:       float32(track.LatitudeInDegrees),
					Lon:       float32(track.LongitudeInDegrees),
					Elevation: int(math.Round(track.AltitudeInMeters)),
					Timestamp: track.Time,
					HeartRate: track.HeartRateInBpm,
				}

				rtc.Points = append(rtc.Points, p)
			}
		}
	}

	return rtc, nil
}
