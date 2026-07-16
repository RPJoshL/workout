package parser

import (
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/models"
)

type intervalConfig struct {
	// minimum speed in sec/km for detecting an interval
	minSpeed int
	// minimum interval to be detected
	minInterval time.Duration
	// Allowed pauses within an interval
	allowedPause time.Duration
}

type interval struct {
	start *models.WorkoutDetails
	end   *models.WorkoutDetails
}

func (i interval) isValid(conf *intervalConfig) bool {
	if i.start == nil || i.end == nil {
		return false
	}

	if i.start.Part != i.end.Part {
		return false
	}

	return time.Duration((i.end.Duration-i.start.Duration)*int(time.Second)) >= conf.minInterval
}

// isLost returns whether the interval was invalideted because of a
// too long break
func (i interval) isLost(current *models.WorkoutDetails, conf *intervalConfig) bool {
	if i.start == nil || i.end == nil {
		return true
	}

	if i.start.Part != i.end.Part {
		return true
	}

	return time.Duration(current.Duration-i.end.Duration)*time.Second > conf.allowedPause
}

func (i interval) toMetric() models.IntervalMetric {
	return models.IntervalMetric{
		From:   i.start.Duration,
		FromID: i.start.Id,
		To:     i.end.Duration,
		ToID:   i.end.Id,
	}
}

func newInterval(start *models.WorkoutDetails) interval {
	return interval{
		start: start,
		end:   start,
	}
}

func GetIntervalMetrics(workout *models.Workout) []models.IntervalMetric {
	config := getIntervalConfig(workout.TypeId)
	if config == nil {
		return nil
	}

	var rtc []models.IntervalMetric

	var lastInterval interval
	for _, point := range workout.WorkoutDetails {
		if point.Speed > config.minSpeed || point.Speed == 0 {
			continue
		}

		if lastInterval.isLost(&point, config) {
			if lastInterval.isValid(config) {
				rtc = append(rtc, lastInterval.toMetric())
			}

			lastInterval = newInterval(&point)
			continue
		}

		lastInterval.end = &point
	}

	if lastInterval.isValid(config) {
		rtc = append(rtc, lastInterval.toMetric())
	}

	return rtc
}

func getIntervalConfig(typeID int) *intervalConfig {
	if typeID == models.TYPE_PUMP_FOILING {
		return &intervalConfig{
			minSpeed:     400, // 9 km/h
			minInterval:  30 * time.Second,
			allowedPause: 12 * time.Second,
		}
	}

	return nil
}
