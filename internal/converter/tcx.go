package converter

import (
	"bytes"
	"encoding/xml"
	"math"
	"strconv"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
	"github.com/RPJoshL/go-logger"
)

var (
	ErrTxcError = errors.BadRequest("#workout.tcxError")

	noGPSPauseThreshold = 120 * time.Second

	garminActivityExtensionNamespace = "http://www.garmin.com/xmlschemas/ActivityExtension/v2"
	garminActivityExtensionLocation  = "https://www8.garmin.com/xmlschemas/ActivityExtensionv2.xsd"
	txcNamespace                     = "http://www.garmin.com/xmlschemas/TrainingCenterDatabase/v2"
	txcSchemaLocation                = "https://www8.garmin.com/xmlschemas/TrainingCenterDatabasev2.xsd"
)

// Tcx is the root struct of a TCX file
type Tcx struct {
	XMLName     xml.Name      `xml:"TrainingCenterDatabase"`
	XSINs       string        `xml:"xmlns:xsi,attr"`
	XMLNs       string        `xml:"xmlns,attr"`
	XMLNSchemas string        `xml:"xsi:schemaLocation,attr"`
	Activities  []TcxActivity `xml:"Activities>Activity"`
}

type TcxActivity struct {
	Sport string   `xml:"Sport,attr"`
	ID    string   `xml:"Id"`
	Laps  []TcxLap `xml:"Lap"`
}

type TcxLap struct {
	StartTime                  time.Time       `xml:"StartTime,attr"`
	TotalTimeInSeconds         float64         `xml:"TotalTimeSeconds"`
	DistanceInMeters           float64         `xml:"DistanceMeters"`
	MaximumSpeedInMetersPerSec float64         `xml:"MaximumSpeed,omitempty"`
	Calories                   float64         `xml:"Calories"`
	Intensity                  string          `xml:"Intensity"`
	TriggerMethod              string          `xml:"TriggerMethod"`
	Track                      []TcxTrackpoint `xml:"Track>Trackpoint"`
	Notes                      string          `xml:"Notes,omitempty"`
	AverageHeartRateBpm        int             `xml:"AverageHeartRateBpm,omitempty"`
	MaximumHeartRateBpm        int             `xml:"MaximumHeartRateBpm,omitempty"`
}

type TcxTrackpoint struct {
	Time             time.Time     `xml:"Time"`
	Position         *TxcPosition  `xml:"Position,omitempty"`
	AltitudeInMeters float64       `xml:"AltitudeMeters"`
	HeartRateInBpm   int           `xml:"HeartRateBpm>Value"`
	Distance         float64       `xml:"DistanceMeters"`
	Extension        TxcExtensions `xml:"Extensions"`
}

type TxcPosition struct {
	LatitudeDegrees  float64 `xml:"LatitudeDegrees"`
	LongitudeDegrees float64 `xml:"LongitudeDegrees"`
}

type TxcExtensions struct {
	XMLName xml.Name                 `xml:"Extensions"`
	TPX     *garminActivityExtension `xml:"TPX,omitempty"`
}

type garminActivityExtension struct {
	Xmlns string  `xml:"xmlns,attr"`
	Speed float64 `xml:"Speed"`
	Steps int     `xml:"Steps,omitempty"`
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
		pauseDuration = int(noGPSPauseThreshold.Seconds())
	}

	// Transform to generic GpxFile
	rtc := &models.GpxFile{}

	// Try to get a workout name
	if len(tcx.Activities) > 0 {
		rtc.Type = models.GetWorkoutTypeByName(tcx.Activities[0].Sport)
	}

	validGPSPoints := 0

	// Collect all Activities and tracks into a single GPX struct
	for _, activity := range tcx.Activities {
		for _, lap := range activity.Laps {
			for _, track := range lap.Track {
				// Basic TCX data
				p := models.GpxPoint{
					Lat:       float32(track.Position.LatitudeDegrees),
					Lon:       float32(track.Position.LongitudeDegrees),
					Elevation: int(math.Round(track.AltitudeInMeters)),
					Timestamp: track.Time,
					HeartRate: track.HeartRateInBpm,
					Distance:  int(math.Round(track.Distance)),
				}

				rtc.Points = append(rtc.Points, p)

				if p.Lat != 0 && p.Lon != 0 {
					validGPSPoints++
				}
			}
		}
	}

	// Increase pause duration and use calculated data when we don't have GPS points
	if validGPSPoints <= 2 {
		rtc.DeviceData = models.DeviceData{
			UseDeviceData: true,
			PauseDuration: pauseDuration,
		}

		return removePauses(rtc), nil
	}

	return rtc, nil
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
			if p.Timestamp.Sub(lastDistinctPoint.Timestamp).Abs() > time.Duration(file.PauseDuration)*time.Second && i-lastDistinctPointIndex > 3 {
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

// ToTCX transforms the provided workout to an TXC file.
//
// Multiple laps are currently not supported. We would have to implement
// them as pauses and calculate the data for them on the fly
func ToTCX(data *models.Workout) ([]byte, error) {
	if len(data.WorkoutDetails) == 0 {
		return nil, nil
	}

	txc := TcxActivity{
		Sport: data.Name,
		ID:    strconv.Itoa(data.Id),
	}

	currentLap := TcxLap{
		StartTime:           data.Start,
		TotalTimeInSeconds:  float64(data.Duration),
		DistanceInMeters:    float64(data.Distance),
		Calories:            float64(data.Calories - data.CaloriesDefault),
		Intensity:           getIntensity(int(data.HeartRateAv.ValueOrZero())),
		AverageHeartRateBpm: int(data.HeartRateAv.ValueOrZero()),
		MaximumHeartRateBpm: int(data.HeartRateMax.ValueOrZero()),
		Notes:               data.Note.String,
		TriggerMethod:       "Manual",
	}

	for _, point := range data.WorkoutDetails {
		currentLap.Track = append(currentLap.Track, getTXCTrackPoint(&point))
	}
	txc.Laps = []TcxLap{currentLap}

	root := Tcx{
		XMLNs:       txcNamespace,
		XSINs:       xsiNamespace,
		XMLNSchemas: txcSchemaLocation + " " + garminActivityExtensionLocation,
		Activities:  []TcxActivity{txc},
	}

	// We return it pretty for the user because we don't expect that performance matters
	output, err := xml.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, err
	}

	return output, nil
}

func getIntensity(avgHeartRate int) string {
	if avgHeartRate < 20 {
		return ""
	}

	switch {
	case avgHeartRate > 110:
		return "Active"
	case avgHeartRate > 90:
		return "Warmup"
	default:
		return "Resting"
	}
}

func getTXCTrackPoint(in *models.WorkoutDetails) TcxTrackpoint {
	rtc := TcxTrackpoint{
		Time:             in.Time,
		Distance:         float64(in.Distance),
		AltitudeInMeters: float64(in.Elevation),
		HeartRateInBpm:   int(in.HeartRate.ValueOrZero()),
	}

	if in.Latitude != 0 && in.Longitude != 0 {
		rtc.Position = &TxcPosition{
			LatitudeDegrees:  float64(in.Latitude),
			LongitudeDegrees: float64(in.Longitude),
		}
	}

	if in.Speed > 0 {
		rtc.Extension.TPX = &garminActivityExtension{
			Xmlns: garminActivityExtensionNamespace,
			Speed: float64(in.Speed),
		}
	}

	return rtc
}
