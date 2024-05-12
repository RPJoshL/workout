package workout

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sync"

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
	if err := sel.Selector(database.ColumnSelector{PointedKeyReference: true, ForeignKeyReference: true}).Run(); err != nil {
		if err.Type() == database.NoRows {
			return nil, ErrWorkoutNotFound
		}
		return nil, err.GetResponse().Log("Failed to query workout", err.GetError(), a)
	}

	rtc.DownsampledDetails = a.DownsamplePoints(&rtc.Workout, 2, 150)

	// Get data per km
	rtc.KmData.Points = a.GetKmStats(&rtc.Workout)
	rtc.KmData.MinSpeed = rtc.KmData.Points[0].Speed
	for _, p := range rtc.KmData.Points {
		if p.Speed > rtc.KmData.MaxSpeed {
			rtc.KmData.MaxSpeed = p.Speed
		}
		if p.Speed < rtc.KmData.MinSpeed {
			rtc.KmData.MinSpeed = p.Speed
		}
	}

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
	var mtx sync.Mutex
	var wg sync.WaitGroup
	for _, w := range rtc.WorkoutData {

		wg.Add(1)
		go func(workout models.Workout) {
			defer wg.Done()

			// Get tooltip content
			buf := new(bytes.Buffer)
			a.GetWorkoutPopup(&workout).Render(context.Background(), buf)

			// Downsample points
			downsampled := a.DownsamplePoints(&w, 20, 2000)
			line := leaflet.Line{
				TooltipContent: buf.String(),
			}
			for _, d := range downsampled {
				line.Points = append(line.Points, leaflet.Point{
					Latitude:  d.Latitude,
					Longitude: d.Longitude,
				})
			}

			mtx.Lock()
			rtc.Lines = append(rtc.Lines, line)
			mtx.Unlock()
		}(w)

	}
	wg.Wait()

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

// DownsampleForGraph downsamples the provided graph into segmants that can easily be
// viewed on graphs
func (aa *Api) DownsampleForGraph(workout *models.Workout, threshold int, getX func(w models.WorkoutDetails) float64, getY func(w models.WorkoutDetails) float64) []models.WorkoutDetails {
	return workout.WorkoutDetails
}

// GetWorkoutTypeById returns the full workout type for the provided
// ID
func (a *Api) GetWorkoutTypeById(id int) models.WorkoutType {
	for _, t := range *a.Types {
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

func (a *Api) GetKmStats(workout *models.Workout) (rtc []WorkoutDetailsPerKmPoint) {

	// Get km steps to calculate the average on
	kmSteps := 1
	if workout.Distance > 50000 {
		kmSteps = 4
	} else if workout.Distance > 15000 {
		kmSteps = 2
	}

	// Duration of the last point
	lastDuration := 0
	lastDistance := 0
	avgCount := 0
	lastDetails := workout.WorkoutDetails[0]
	// Current point to add
	currentKm := WorkoutDetailsPerKmPoint{}

	// Calculate things
	for i, d := range workout.WorkoutDetails {

		// New max heartrate
		if d.HeartRate.Int64 > int64(currentKm.MaxHeartrate) {
			currentKm.MaxHeartrate = int(d.HeartRate.Int64)
		}

		// Calculate average heartrate
		timePast := d.Duration - lastDetails.Duration
		if timePast <= 6 {
			for i := 1; i <= int(timePast); i++ {
				avgCount++

				val := float64(d.HeartRate.Int64)
				currentKm.AvgHeartrate += (val - float64(currentKm.AvgHeartrate)) / float64(avgCount)
			}
		} else {
			// Draw a vector between last and current point and calculate value at specific time
			stepsBasis := float64(d.HeartRate.Int64 - lastDetails.HeartRate.Int64)
			step := stepsBasis / float64(timePast)
			for i := 1; i <= int(timePast); i++ {
				avgCount++
				val := step*float64(i) + float64(lastDetails.HeartRate.Int64)
				currentKm.AvgHeartrate += (val - currentKm.AvgHeartrate) / float64(avgCount)
			}
		}

		// New km
		if i == len(workout.WorkoutDetails)-1 || d.Distance >= ((len(rtc)+1)*kmSteps*1000) {
			lastKmInMeters := len(rtc) * kmSteps * 1000

			// Fill header
			if i == len(workout.WorkoutDetails)-1 {
				currentKm.KmDescription = fmt.Sprintf("~%d m", d.Distance-lastKmInMeters)
			} else {
				currentKm.KmDescription = fmt.Sprintf("%d-%d km", len(rtc)*kmSteps, (len(rtc)+1)*kmSteps)
			}

			// Calculate data
			speed := float64(d.Duration-lastDuration) / (float64(d.Distance-lastDistance) / 1000.0)
			currentKm.Speed = int(math.Round(speed))

			// Append to return value
			rtc = append(rtc, currentKm)

			// Reset values
			avgCount = 0
			currentKm = WorkoutDetailsPerKmPoint{}
			lastDuration = d.Duration
			lastDistance = d.Distance
		}

		lastDetails = d
	}

	return rtc
}
