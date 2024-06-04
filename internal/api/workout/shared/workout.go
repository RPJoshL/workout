package shared

import (
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/tkrajina/gpxgo/gpx"
)

// DownsamplePoints downsamples the GPX points of a workout by applying
// the Ramer-Douglas-Peucker algorithm.
//
// All Points have to be inside the provided "toleranz" in meters.
// If two downsampled points are more fare away than "maxPointDistance", the point
// at that distance is added.
// Note that an offset of 20% is used to not draw points directly behind each other
func (a *Shared) DownsamplePoints(workout *models.Workout, toleranz float64, maxPointDistance int) (rtc []models.WorkoutDetails) {

	// No data to transform
	if len(workout.WorkoutDetails) == 0 {
		return
	}

	// Transform workout details into a GPX file required for gpxgo
	segment := gpx.GPXTrackSegment{}
	for _, p := range workout.WorkoutDetails {
		segment.Points = append(segment.Points, gpx.GPXPoint{
			Point: gpx.Point{
				Latitude:  p.Latitude,
				Longitude: p.Longitude,
				Elevation: *gpx.NewNullableFloat64(float64(p.Elevation)),
			},
		})
	}

	// Simplify this gpx file with a max offset distance of 2 meters
	gpxFile := gpx.GPX{
		Tracks: []gpx.GPXTrack{{Segments: []gpx.GPXTrackSegment{segment}}},
	}
	gpxFile.SimplifyTracks(toleranz)
	newPoints := gpxFile.Tracks[0].Segments[0].Points

	// Find the matching workout details to the downsampled points
	maxDistanceThreshold := float64(maxPointDistance) * 1.2
	iDetails := 0
	lastPoint := workout.WorkoutDetails[0]
	for _, p := range newPoints {

		// Distance how far this point is away from the last point
		pointDistance := gpx.Distance2D(lastPoint.Latitude, lastPoint.Longitude, p.Latitude, p.Longitude, false)

		// Find this point in existing workout details
		found := false
		for ; iDetails < len(workout.WorkoutDetails); iDetails++ {
			dd := workout.WorkoutDetails[iDetails]

			// If the last point is more far away than the threshold, add a temporary
			// point that is not downsampled
			if pointDistance > maxDistanceThreshold && dd.Distance-lastPoint.Distance > maxPointDistance {
				rtc = append(rtc, dd)
				lastPoint = dd
				pointDistance = gpx.Distance2D(lastPoint.Latitude, lastPoint.Longitude, p.Latitude, p.Longitude, false)
			}

			// We found the edge point
			if dd.Latitude == p.Latitude && dd.Longitude == p.Longitude && dd.Elevation == int(p.Elevation.Value()) {
				rtc = append(rtc, dd)
				found = true
				break
			}
		}

		if !found {
			a.Logger().Warning("Didn't found a downsampled point for: LAT %f | LON %f | ELE %f", p.Latitude, p.Longitude, p.Elevation.Value())
		}
	}

	a.Logger().Trace("Downsampled points from %d to %d", len(workout.WorkoutDetails), len(rtc))
	return
}

// DownsampleForGraph downsamples the provided graph into segmants that can easily be
// viewed on graphs
func (aa *Shared) DownsampleForGraph(workout *models.Workout, threshold int, getX func(w models.WorkoutDetails) float64, getY func(w models.WorkoutDetails) float64) []models.WorkoutDetails {
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
