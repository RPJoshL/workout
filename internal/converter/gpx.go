package converter

import (
	"math"
	"strconv"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/tkrajina/gpxgo/gpx"
)

var (
	ErrGpxError = errors.BadRequest("#workout.gpxError")
)

type gpxW struct {
	content *[]byte
	gpx     *gpx.GPX
}

func ParseGPX(content []byte) (*models.GpxFile, errors.Error) {
	// Parse file
	gpx, err := gpx.ParseBytes(content)
	if err != nil {
		logger.Error("Unable to decode TCX file: %s", err)
		return nil, ErrGpxError
	}

	// Remove extremes
	gpx.RemoveHorizontalExtremes()
	gpx.RemoveVerticalExtremes()

	gpo := &gpxW{
		gpx:     gpx,
		content: &content,
	}

	// Transform to generic GpxFile
	rtc := &models.GpxFile{}

	// Try to determine workout name
	rtc.Type = gpo.getWorkoutType()

	// Collect all tracks into a single workout
	for tIdx := range gpx.Tracks {
		// Parse all segments and points
		for _, segment := range gpx.Tracks[tIdx].Segments {
			for i := range segment.Points {
				if p, e := transformGpxPoint(&segment.Points[i]); e != nil {
					return nil, e
				} else {
					rtc.Points = append(rtc.Points, p)
				}
			}
		}
	}

	// Fall back to simple waypoints
	if len(rtc.Points) == 0 {
		for i := range gpx.Waypoints {
			if p, e := transformGpxPoint(&gpx.Waypoints[i]); e != nil {
				return nil, e
			} else {
				rtc.Points = append(rtc.Points, p)
			}
		}
	}

	return rtc, nil
}

// transformGpxPoint transforms the provided GPX point into our generic representation
func transformGpxPoint(point *gpx.GPXPoint) (models.GpxPoint, errors.Error) {
	var err error

	// Basic GPX data
	p := models.GpxPoint{
		Lat:       float32(point.Latitude),
		Lon:       float32(point.Longitude),
		Elevation: int(math.Round(point.Elevation.Value())),
		Timestamp: point.Timestamp,
	}

	// Parse garmin trackpoint extension
	if trackPointExt, found := point.Extensions.GetNode("http://www.garmin.com/xmlschemas/TrackPointExtension/v1", "TrackPointExtension"); found {
		for _, ext := range trackPointExt.Nodes {
			if ext.XMLName.Local == "hr" {
				heartRateStr := ext.Data
				p.HeartRate, err = strconv.Atoi(heartRateStr)
				if err != nil {
					logger.Error("Failed to convert heart rate of TrackPointExtension with value %q: %s", heartRateStr, err)
					return p, ErrGpxError
				}
			}
		}
	}

	return p, nil
}

// getName returns the workouts name specified within
// the GPX file
func (g *gpxW) getWorkoutType() int {
	// Global name attribute
	if val := models.GetWorkoutTypeByName(g.gpx.Name); val != models.TYPE_UNKNOWN {
		return val
	}

	// First track segment
	if len(g.gpx.Tracks) > 0 {
		if val := models.GetWorkoutTypeByName(g.gpx.Tracks[0].Name); val != models.TYPE_UNKNOWN {
			return val
		}
	}

	return models.TYPE_UNKNOWN
}
