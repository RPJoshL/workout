package converter

import (
	"fmt"
	"math"
	"strconv"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/tkrajina/gpxgo/gpx"
)

type gpxW struct {
	content *[]byte
	gpx     *gpx.GPX
}

func ParseGPX(content []byte) (*models.GpxFile, error) {

	// Parse file
	gpx, err := gpx.ParseBytes(content)
	if err != nil {
		return nil, err
	}
	gpo := &gpxW{
		gpx:     gpx,
		content: &content,
	}

	// Transform to generic GpxFile
	rtc := &models.GpxFile{}

	// Try to determine workout name
	rtc.Type = gpo.getWorkoutType()

	// Collect all tracks into a single workout
	for _, track := range gpx.Tracks {
		// Parse all segments and points
		for _, segment := range track.Segments {
			for _, point := range segment.Points {

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
								return nil, fmt.Errorf("failed to convert heart rate of TrackPointExtension with value %q: %s", heartRateStr, err)
							}
						}
					}
				}

				rtc.Points = append(rtc.Points, p)
			}
		}
	}

	return rtc, nil
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
