package models

import (
	"fmt"
	"math"

	"github.com/guregu/null/v5"
)

var metricTransformer = []MetricTransformerDb{
	&IntervalMetricTransformer{},
}

type MetricType string

const (
	IntervalMetricType MetricType = "IntervalMetric"
)

// MetricTransformerDb transforms a specific metric
// into/from the DB format
type MetricTransformerDb interface {
	// FromDB should set the specific field inside [Workout] after
	// transforming the metrics
	FromDB(workout *Workout, typemetrics []WorkoutMetric)
	ToDB(workout *Workout) []WorkoutMetric
}

type IntervalMetricTransformer struct{}

func (i *IntervalMetricTransformer) FromDB(workout *Workout, typemetrics []WorkoutMetric) {
	workout.IntervalMetric = make([]IntervalMetric, 0, len(typemetrics))

	for _, metric := range typemetrics {
		fromDetails, fromIdx := findWorkoutDetailsByDuration(int(metric.IntVal1.Int64), int(metric.IntVal3.Int64), workout)
		toDetails, toIdx := findWorkoutDetailsByDuration(int(metric.IntVal2.Int64), int(metric.IntVal3.Int64), workout)

		if fromDetails == nil || toDetails == nil || toIdx < fromIdx {
			continue
		}

		lastDetails := workout.WorkoutDetails[fromIdx]
		maxHeartrate, avgHeartrate, cnt := 0, 0.0, 0
		for i := fromIdx; i <= toIdx; i++ {
			details := workout.WorkoutDetails[i]

			if details.HeartRate.Int64 > int64(maxHeartrate) {
				maxHeartrate = int(details.HeartRate.Int64)
			}

			timePast := details.Duration - lastDetails.Duration
			if timePast <= 6 {
				for i := 1; i <= timePast; i++ {
					cnt++

					val := float64(details.HeartRate.Int64)
					avgHeartrate += (val - float64(avgHeartrate)) / float64(cnt)
				}
			} else {
				// Draw a vector between last and current point and calculate value at specific time
				stepsBasis := float64(details.HeartRate.Int64 - lastDetails.HeartRate.Int64)
				step := stepsBasis / float64(timePast)
				for i := 1; i <= timePast; i++ {
					cnt++
					val := step*float64(i) + float64(lastDetails.HeartRate.Int64)
					avgHeartrate += (val - avgHeartrate) / float64(cnt)
				}
			}

			lastDetails = details
		}

		distance := toDetails.Distance - fromDetails.Distance
		duration := toDetails.Duration - fromDetails.Duration
		speed := 0.0

		if distance > 0 {
			speed = float64(duration) / (float64(distance) / 1000.0)

			// Cut the max speed at 30 min/km
			if speed > (30 * 60) {
				speed = 0
			}
		}

		workout.IntervalMetric = append(workout.IntervalMetric, IntervalMetric{
			From:         fromDetails.Duration,
			To:           toDetails.Duration,
			FromID:       fromDetails.Id,
			ToID:         toDetails.Id,
			MaxHeartrate: maxHeartrate,
			AvgHeartrate: int(avgHeartrate),
			AvgSpeed:     int(math.Round(speed)),
			Duration:     duration,
			Part:         fromDetails.Part,
		})
	}
}

func (i *IntervalMetricTransformer) ToDB(workout *Workout) []WorkoutMetric {
	metrics := make([]WorkoutMetric, 0, len(workout.IntervalMetric))

	partDurations := map[int]int{
		0: 0,
	}

	for _, metric := range workout.IntervalMetric {
		partDuration, ok := partDurations[metric.Part]
		if !ok {
			for _, details := range workout.WorkoutDetails {
				if details.Part == metric.Part {
					partDuration = details.Duration
					partDurations[metric.Part] = partDuration
					break
				}
			}
		}

		metrics = append(metrics, WorkoutMetric{
			WorkoutId: workout.Id,
			Type:      string(IntervalMetricType),
			IntVal1:   newInt(metric.From - partDuration),
			IntVal2:   newInt(metric.To - partDuration),
			IntVal3:   newInt(metric.Part),
		})
	}

	return metrics
}

func newInt(val int) null.Int64 {
	return null.IntFrom(int64(val))
}

type IntervalMetric struct {
	// Duration in seconds
	From int `json:"from"`
	// ID of the workout details
	FromID int `json:"from_id"`
	To     int `json:"to"`
	// ID of the workout details
	ToID         int `json:"to_id"`
	AvgHeartrate int `json:"avg_heartrate"`
	MaxHeartrate int `json:"max_heartrate"`
	// Average speed in sec/km
	AvgSpeed int `json:"avg_speed"`
	// Duration of this interval in seconds
	Duration int `json:"duration"`
	// Part of the workout this interval belongs to
	Part int `json:"part"`
}

func findWorkoutDetailsByDuration(duration, part int, workout *Workout) (details *WorkoutDetails, idx int) {
	partStartDuration := -1

	for idx := range workout.WorkoutDetails {
		if workout.WorkoutDetails[idx].Part != part {
			continue
		}

		if partStartDuration == -1 {
			partStartDuration = workout.WorkoutDetails[idx].Duration
		}

		partDuration := workout.WorkoutDetails[idx].Duration - partStartDuration

		if partDuration == duration {
			return &workout.WorkoutDetails[idx], idx
		}
	}

	return nil, 0
}

func (i *IntervalMetric) FormatDuration() string {
	duration := i.To - i.From

	minutes := duration / 60
	seconds := duration % 60

	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func (i *IntervalMetric) GetStartDuration() string {
	return formatDurationShort(i.From)
}

func (i *IntervalMetric) GetEndDuration() string {
	return formatDurationShort(i.To)
}

func formatDurationShort(duration int) string {
	if duration >= (60 * 60) {
		return fmt.Sprintf("%d:%02d:%02d", duration/3600, (duration/60)%60, duration%60)
	} else {
		return fmt.Sprintf("%d:%02d", duration/60, duration%60)
	}
}

// FormatSpeed returns the speed in km/h
func (i *IntervalMetric) FormatSpeed() string {
	if i.AvgSpeed == 0 {
		return "-"
	}

	inKmPerHour := 3600 / float64(i.AvgSpeed)

	return fmt.Sprintf("%.1f km/h", inKmPerHour)
}
