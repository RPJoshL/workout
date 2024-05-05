package workout

import (
	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/go-webserver/errors"
	"git.rpjosh.de/RPJosh/workout/internal/api/components/leaflet"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/tkrajina/gpxgo/gpx"
)

var (
	ErrWorkoutNotFound = errors.NewError("#workout.notFound", 404)
)

// GetWorkoutDetailsData returns the workout data for a specific workout
// identified by the provided ID
func (a *Api) GetWorkoutDetailsData(id int) (*WorkouDetails, errors.Error) {
	rtc := &WorkouDetails{}

	// Get workout
	sel := a.R().Db.Struct.Query(&rtc.Workout)
	sel.Where().Column(models.Workout_UserId, "=", a.R().User.Id).Add()
	sel.Where().Column(models.Workout_Id, "=", id).Add()
	if err := sel.Selector(database.ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		if err.Type() == database.NoRows {
			return nil, ErrWorkoutNotFound
		}
		return nil, err.GetResponse().Log("Failed to query workout", err.GetError(), a)
	}

	rtc.Workout.WorkoutDetails = a.DownsamplePoints(&rtc.Workout, 2, 300)
	return rtc, nil
}

// GetTableData returns workout data for the overview table based
// on the provided search values
func (a *Api) GetTableData() (*TableData, errors.Error) {
	rtc := &TableData{}

	// Get filtered workouts
	sel := a.R().Db.Struct.QuerySlice(&rtc.WorkoutData)
	sel.Where().Column(models.Workout_UserId, "=", a.R().User.Id).Add()
	if err := sel.Selector(database.ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		return nil, err.GetResponse().Log("Failed to query workout", err.GetError(), a)
	}

	// Get downsampled workout data
	for _, w := range rtc.WorkoutData {
		downsampled := a.DownsamplePoints(&w, 20, 2000)
		line := leaflet.Line{
			TooltipContent: "Hello World!",
		}
		for _, d := range downsampled {
			line.Points = append(line.Points, leaflet.Point{
				Latitude:  d.Latitude,
				Longitude: d.Longitude,
			})
		}

		rtc.Lines = append(rtc.Lines, line)
	}

	return rtc, nil
}

// DownsamplePoints downsamples the GPX points of a workout by applying
// the Ramer-Douglas-Peucker algorithm.
//
// All Points have to be inside the provided "toleranz" in meters.
// If two downsampled points are more fare away than "maxPointDistance", the point
// at that distance is added.
// Note that an offset of 20% is used to not draw points directly behind each other
func (a *Api) DownsamplePoints(workout *models.Workout, toleranz float64, maxPointDistance int) (rtc []models.WorkoutDetails) {

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
			logger.Warning("Didn't found a downsampled point for: LAT %f | LON %f | ELE %f", p.Latitude, p.Longitude, p.Elevation.Value())
		}
	}

	logger.Trace("Downsampled points from %d to %d", len(workout.WorkoutDetails), len(rtc))
	return
}
