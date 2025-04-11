package details

import (
	"fmt"
	"math"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/database/dbstruct"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

var (
	ErrWorkoutNotFound = errors.NewError("#workout.notFound", 404)
	ErrTime            = errors.NewError("Invalid date format provided: %q", 400)
)

// GetWorkoutDetailsData returns the workout data for a specific workout
// identified by the provided ID
func (api *Api) GetWorkoutDetailsData(id int) (*WorkouDetails, errors.Error) {
	rtc := &WorkouDetails{}

	// Get workout
	sel := api.R().Db.Struct.Query(&rtc.Workout)
	sel.Where().Column(models.Workout_UserId, "=", api.R().User.Id).Add()
	sel.Where().Column(models.Workout_Id, "=", id).Add()
	sel.OrderBy("workout_details", models.WorkoutDetails_Duration, "ASC")
	if err := sel.Selector(dbstruct.ColumnSelector{PointedKeyReference: true, ForeignKeyReference: true}).Run(); err != nil {
		if err.Type() == database.NoRows {
			return nil, ErrWorkoutNotFound
		}
		return nil, err.GetResponse().Log("Failed to query workout", err.GetError(), api)
	}

	rtc.DownsampledDetails = api.Shared.DownsamplePoints(&rtc.Workout, 2, 150)

	// We cannot do anything if we don't have any points
	if len(rtc.Workout.WorkoutDetails) == 0 {
		return rtc, nil
	}

	// Get data per km
	rtc.KmData.Points = api.GetKmStats(&rtc.Workout)
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

func (api *Api) GetKmStats(workout *models.Workout) (rtc []WorkoutDetailsPerKmPoint) {
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
			for i := 1; i <= timePast; i++ {
				avgCount++

				val := float64(d.HeartRate.Int64)
				currentKm.AvgHeartrate += (val - float64(currentKm.AvgHeartrate)) / float64(avgCount)
			}
		} else {
			// Draw a vector between last and current point and calculate value at specific time
			stepsBasis := float64(d.HeartRate.Int64 - lastDetails.HeartRate.Int64)
			step := stepsBasis / float64(timePast)
			for i := 1; i <= timePast; i++ {
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
