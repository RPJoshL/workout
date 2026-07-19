package parser

import (
	"math"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/RPJoshL/go-logger"
)

// PostProcessor is responsible for applying post processing steps to the already
// parsed workout data
type PostProcessor struct {
	opt PostProcessingOptions
}

type PostProcessingOptions struct {
	UseSpeedDeviceData bool
}

func NewPostProcessor(opt PostProcessingOptions) *PostProcessor {
	return &PostProcessor{
		opt: opt,
	}
}

// PostProcess applies any post processing steps to the workout data
// in order to remove "errors" from the tracking device:
//   - Usage of an average speed for missing GPS data
func (p *PostProcessor) PostProcess(workout *models.Workout) {
	// The tracking device has to implement this.
	// And we don't know how the data was tracked, so we cannot do any assumptions about the data quality
	if !p.opt.UseSpeedDeviceData {
		// Index of the last details which still had a valid speed value
		lastSpeedIdx := -1

		for idx, point := range workout.WorkoutDetails {
			if isValidSpeed(point.Speed) {
				p.handleInvalidSpeed(lastSpeedIdx, idx, workout)
				lastSpeedIdx = idx
			}
		}
	}

	workout.IntervalMetric = GetIntervalMetrics(workout)
}

// handleInvalidSpeed handles and modifies missing speed values.
// "from" and "to" should be points where a speed value was lastly detected
func (p *PostProcessor) handleInvalidSpeed(from, to int, workout *models.Workout) {
	diff := to - from
	if from < 0 || (from+1) == to || diff > 7 {
		return
	}

	firstPoint := workout.WorkoutDetails[from]
	lastPoint := workout.WorkoutDetails[to]

	// Lastpoint should have a much higher speed
	avgThreshold := 0.55
	if diff > 4 {
		avgThreshold = 0.35
	}
	if float64(lastPoint.Speed) > (float64(workout.SpeedAv)*avgThreshold) || float64(lastPoint.Speed) > (float64(firstPoint.Speed)*0.7) {
		return
	}

	// Validate that a location was missing which resulted into incorrect speed values
	for i := from + 1; i < to; i++ {
		point := workout.WorkoutDetails[i]

		sameGps := point.Latitude == firstPoint.Latitude || point.Longitude == firstPoint.Longitude
		if isValidSpeed(point.Speed) || !sameGps || point.Part != firstPoint.Part {
			return
		}
	}

	avgSpeedF := float64(lastPoint.Duration-firstPoint.Duration) / (float64(lastPoint.Distance-firstPoint.Distance) / 1000)

	avgSpeed := int(math.Round(avgSpeedF))
	if !isValidSpeed(avgSpeed) {
		return
	}
	logger.Debug("Setting average speed for %d - %d (in %d)", from, to, workout.Id)

	for i := from + 1; i <= to; i++ {
		point := &workout.WorkoutDetails[i]
		point.Speed = avgSpeed
	}
}

func isValidSpeed(speed int) bool {
	return speed > 0 && speed < 4000
}
