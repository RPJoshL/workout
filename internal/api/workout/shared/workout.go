package shared

import (
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/tkrajina/gpxgo/gpx"
)

type DownSampleConstraints struct {
	// Maximum distance in meters how far a point can be away from the original track
	MaxPointDistance int
	// Maximum duration in seconds how far a point can be away from the last one
	MaxDuration int

	// The downsampling is driven by the last constraint that was exceeded. So points do have always the same distance / duration to each other
	// and points defined by the algorithm are added in addition to these constraints
	ConstraintDriven bool
}

// isAdditionalExceeded checks if the provided point exceeds the defined constraints.
// As a result, it should be added to the downsampled points as well, even if it is not defined by the Ramer-Douglas-Peucker algorithm
func (c DownSampleConstraints) addAdditionalPoint(lastPoint, currentPoint *models.WorkoutDetails, downsampled *gpx.GPXPoint) bool {
	if c.MaxPointDistance > 0 {
		distanceToDownsampled := gpx.Distance2D(lastPoint.Latitude, lastPoint.Longitude, downsampled.Latitude, downsampled.Longitude, false)
		distanceToCurrent := gpx.Distance2D(lastPoint.Latitude, lastPoint.Longitude, currentPoint.Latitude, currentPoint.Longitude, false)

		// Use a threshold of 15% to not draw points directly behind each other
		if distanceToDownsampled > float64(c.MaxPointDistance)*1.15 && distanceToCurrent > float64(c.MaxPointDistance) {
			return true
		}
	}

	if c.MaxDuration > 0 {
		durationToCurrent := currentPoint.Duration - lastPoint.Duration

		if durationToCurrent >= c.MaxDuration {
			return true
		}
	}

	return false
}

// DownsamplePoints downsamples the GPX points of a workout by applying
// the Ramer-Douglas-Peucker algorithm.
//
// All Points have to be inside the provided "toleranz" in meters.
// Additional points besides the points defined by the algorithm are added if the constraints are exceeded
func (a *Shared) DownsamplePoints(workout *models.Workout, toleranz float64, constraints DownSampleConstraints) (rtc []models.WorkoutDetails) {
	// No data to transform
	if len(workout.WorkoutDetails) == 0 {
		return
	}

	iDetails := 0
	lastPoint := workout.WorkoutDetails[0]
	rtc = append(rtc, lastPoint)

	simplified := a.simplify(workout, toleranz)
	for i := range simplified {
		downSampled := &simplified[i]
		downSampledDetails := &models.WorkoutDetails{}

		// Find the downsampled point in the original workout details
		downI := iDetails
		found := false
		for ; downI < len(workout.WorkoutDetails); downI++ {
			dd := &workout.WorkoutDetails[downI]
			if dd.Time.Equal(downSampled.Timestamp) {
				downSampledDetails = dd
				found = true
				break
			}
		}

		if !found {
			a.Logger().Warning("Didn't found a downsampled point for: LAT %f | LON %f | ELE %f", downSampled.Latitude, downSampled.Longitude, downSampled.Elevation.Value())
			continue
		}

		for ; iDetails <= downI; iDetails++ {
			dd := workout.WorkoutDetails[iDetails]

			// Add a temporary point that is not downsampled if the constraints are exceeded
			if constraints.addAdditionalPoint(&lastPoint, &dd, downSampled) {
				rtc = append(rtc, dd)
				lastPoint = dd
			}
		}

		// Add downsampled point
		if lastPoint.Duration != downSampledDetails.Duration {
			rtc = append(rtc, *downSampledDetails)
		}
		if !constraints.ConstraintDriven {
			lastPoint = *downSampledDetails
		}
	}

	a.Logger().Trace("Downsampled points from %d to %d", len(workout.WorkoutDetails), len(rtc))
	return
}

// simplify simplifies the provided workout points with [gpx.GpxFile.SimplifyTracks]
func (a *Shared) simplify(workout *models.Workout, toleranz float64) (rtc []gpx.GPXPoint) {
	// Transform workout details into a GPX file (with segments) required for gpxgo
	segments := []gpx.GPXTrackSegment{}
	currentSegment := gpx.GPXTrackSegment{}
	lastSegmentIndex := workout.WorkoutDetails[0].Part
	for _, p := range workout.WorkoutDetails {
		if lastSegmentIndex != p.Part {
			segments = append(segments, currentSegment)
			currentSegment = gpx.GPXTrackSegment{}
			lastSegmentIndex = p.Part
		}

		currentSegment.Points = append(currentSegment.Points, gpx.GPXPoint{
			Timestamp: p.Time,
			Point: gpx.Point{
				Latitude:  p.Latitude,
				Longitude: p.Longitude,
				Elevation: *gpx.NewNullableFloat64(float64(p.Elevation)),
			},
		})
	}
	segments = append(segments, currentSegment)

	// Simplify this gpx file with a max offset distance of 2 meters
	gpxFile := gpx.GPX{
		Tracks: []gpx.GPXTrack{{Segments: segments}},
	}
	gpxFile.SimplifyTracks(toleranz)

	for _, p := range gpxFile.Tracks[0].Segments {
		rtc = append(rtc, p.Points...)
	}

	return
}

// DownsampleForGraph downsamples the provided graph into segments that can easily be
// viewed on graphs
func (a *Shared) DownsampleForGraph(workout *models.Workout, threshold int, getX, getY func(w models.WorkoutDetails) float64) []models.WorkoutDetails {
	return workout.WorkoutDetails
}

// GetWorkoutTypeById returns the full workout type for the provided
// ID
func (a *Shared) GetWorkoutTypeById(id int) models.WorkoutType {
	for _, t := range WorkoutTypes {
		if t.Id == id {
			return t
		}
	}

	return models.WorkoutType{
		Id:     -1,
		NameDe: "Unbekannt",
		NameEn: "Unknown",
	}
}
