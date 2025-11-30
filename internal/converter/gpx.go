package converter

import (
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/tkrajina/gpxgo/gpx"
)

var (
	ErrGpxError = errors.BadRequest("#workout.gpxError")

	gpxNamespace = "http://www.topografix.com/GPX/1/1"
	gpxSchemaLoc = "http://www.topografix.com/GPX/1/1/gpx.xsd"

	trackPointExtensionNamespace = "http://www.garmin.com/xmlschemas/TrackPointExtension/v1"
	trackPointExtensionLocation  = "https://www8.garmin.com/xmlschemas/TrackPointExtensionv1.xsd"

	xsiNamespace = "http://www.w3.org/2001/XMLSchema-instance"
)

type gpxW struct {
	content *[]byte
	gpx     *gpx.GPX
}

type gpxFile struct {
	XMLName             xml.Name       `xml:"gpx"`
	XMLNs               string         `xml:"xmlns,attr"`
	XSINs               string         `xml:"xmlns:xsi,attr"`
	XmlSchemaLoc        string         `xml:"xsi:schemaLocation,attr"`
	TrackPointNamespace string         `xml:"xmlns:gpxtpx,attr"`
	Creator             string         `xml:"creator,attr"`
	Name                string         `xml:"name,omitempty"`
	Version             string         `xml:"version,attr,omitempty"`
	Tracks              []*gpx11GpxTrk `xml:"trk"`
}

type gpx11GpxTrk struct {
	XMLName  xml.Name          `xml:"trk"`
	Name     string            `xml:"name,omitempty"`
	Number   int               `xml:"number"`
	Type     string            `xml:"type,omitempty"`
	Segments []*gpx11GpxTrkSeg `xml:"trkseg,omitempty"`
}

type gpx11GpxTrkSeg struct {
	XMLName xml.Name         `xml:"trkseg"`
	Points  []*gpx11GpxPoint `xml:"trkpt"`
}

type gpx11GpxPoint struct {
	Lat        float64         `xml:"lat,attr"`
	Lon        float64         `xml:"lon,attr"`
	Ele        int             `xml:"ele,omitempty"`
	Timestamp  string          `xml:"time,omitempty"`
	Type       string          `xml:"type,omitempty"`
	Extensions *gpx11Extension `xml:"extensions,omitempty"`
}

type gpx11Extension struct {
	TrackPointExtension *trackPointExtension `xml:"gpxtpx:TrackPointExtension,omitempty"`
}

type trackPointExtension struct {
	HR int `xml:"gpxtpx:hr,omitempty"`
}

func ParseGPX(content []byte) (*models.GpxFile, errors.Error) {
	// Parse file
	file, err := gpx.ParseBytes(content)
	if err != nil {
		logger.Error("Unable to decode TCX file: %s", err)
		return nil, ErrGpxError
	}

	// Remove extremes
	file.RemoveHorizontalExtremes()
	file.RemoveVerticalExtremes()

	gpo := &gpxW{
		gpx:     file,
		content: &content,
	}

	// Transform to generic GpxFile
	rtc := &models.GpxFile{}

	// Try to determine workout name
	rtc.Type = gpo.getWorkoutType()

	// Collect all tracks into a single workout
	for tIdx := range file.Tracks {
		// Parse all segments and points
		for _, segment := range file.Tracks[tIdx].Segments {
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
		for i := range file.Waypoints {
			if p, e := transformGpxPoint(&file.Waypoints[i]); e != nil {
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

// ToGPX transforms the provided workout into an GPX file
func ToGPX(in *models.Workout) ([]byte, error) {
	file := gpxFile{
		XMLNs:               gpxNamespace,
		XSINs:               xsiNamespace,
		XmlSchemaLoc:        gpxSchemaLoc + " " + trackPointExtensionLocation,
		TrackPointNamespace: trackPointExtensionNamespace,
		Creator:             "RPout",
		Name:                in.Name,
		Version:             "1.1",
		Tracks:              getGPXTracks(in),
	}

	rtc, err := xml.MarshalIndent(file, "", "  ")
	if err != nil {
		return []byte{}, err
	}

	return rtc, nil
}

func getGPXTracks(in *models.Workout) []*gpx11GpxTrk {
	if len(in.WorkoutDetails) == 0 {
		return []*gpx11GpxTrk{}
	}

	rtc := []*gpx11GpxTrk{}

	currentTrack := newGPXTrack(&in.WorkoutDetails[0])
	for _, p := range in.WorkoutDetails {
		// Init a new track
		if p.Part != currentTrack.Number && len(currentTrack.Segments[0].Points) > 0 {
			rtc = append(rtc, currentTrack)

			currentTrack = newGPXTrack(&p)
		}

		currentTrack.Segments[0].Points = append(currentTrack.Segments[0].Points, toGPXPoint(&p))
	}

	if len(currentTrack.Segments[0].Points) > 0 {
		rtc = append(rtc, currentTrack)
	}

	return rtc
}

func newGPXTrack(p *models.WorkoutDetails) *gpx11GpxTrk {
	return &gpx11GpxTrk{
		Number: p.Part,
		Name:   fmt.Sprintf("%d. Part", p.Part+1),
		Segments: []*gpx11GpxTrkSeg{
			{},
		},
	}
}

func toGPXPoint(p *models.WorkoutDetails) *gpx11GpxPoint {
	rtc := &gpx11GpxPoint{
		Lat:       p.Latitude,
		Lon:       p.Longitude,
		Ele:       p.Elevation,
		Timestamp: p.Time.Format(time.RFC3339),
	}

	if p.HeartRate.Valid {
		rtc.Extensions = &gpx11Extension{
			TrackPointExtension: &trackPointExtension{
				HR: int(p.HeartRate.Int64),
			},
		}
	}

	return rtc
}
