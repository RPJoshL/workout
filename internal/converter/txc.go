package converter

import (
	"bytes"
	"encoding/xml"
	"math"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
)

var (
	ErrTxcError = errors.BadRequest("#workout.tcxError")

	defaultPauseThreshold = 120 * time.Second
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
	Distance           float64   `xml:"DistanceMeters"`
}

// ParseTcx parses the file content of a TCX file and return the
// transformed GPX struct
func ParseTcx(content []byte, pauseDuration int) (*models.GpxFile, errors.Error) {
	// Parse the file
	tcx := Tcx{}
	decoder := xml.NewDecoder(bytes.NewBuffer(content))
	if err := decoder.Decode(&tcx); err != nil {
		logger.Error("Unable to decode TCX file: %s", err)
		return nil, ErrTxcError
	}

	if pauseDuration <= 5 {
		pauseDuration = int(defaultPauseThreshold.Seconds())
	}

	// Transform to generic GpxFile
	rtc := &models.GpxFile{
		DeviceData: models.DeviceData{
			// Because we don't have GPS points
			UseDeviceData: true,
			PauseDuration: pauseDuration,
		},
	}

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
					Distance:  int(math.Round(track.Distance)),
				}

				rtc.Points = append(rtc.Points, p)
			}
		}
	}

	return removePauses(rtc), nil
}

// removePauses removes any pauses in a TXC / GPX file based on
// identicall values for more than 2 minutes
func removePauses(file *models.GpxFile) *models.GpxFile {
	if len(file.Points) < 5 {
		return file
	}

	// The last point that was different from the previous one
	lastDistinctPoint := file.Points[0]
	lastDistinctPointIndex := 0

	for i := 0; i < len(file.Points); i++ {
		p := file.Points[i]

		if !lastDistinctPoint.EqualValues(p) {
			// If the gap is bigger than two minutes (and we have multiple points), set them paused until the end
			if p.Timestamp.Sub(lastDistinctPoint.Timestamp).Abs() > time.Duration(file.DeviceData.PauseDuration)*time.Second && i-lastDistinctPointIndex > 3 {
				logger.Trace("Detected a pause in file from %q to %q (%d - %d)", lastDistinctPoint.Timestamp.Format(time.RFC3339), p.Timestamp.Format(time.RFC3339), lastDistinctPointIndex, i)
				// Remove all points starting after the last distinct point
				iReal := i
				for a := lastDistinctPointIndex + 1; a < iReal; a++ {
					file.Points = utils.RemovePreserveOrder(&file.Points, i-1)
					i--
				}
			}

			// Different point → update reference
			lastDistinctPoint = file.Points[i]
			lastDistinctPointIndex = i
		}
	}

	return file
}
