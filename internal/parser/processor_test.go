package parser

import (
	"fmt"
	"testing"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/assert"
)

// TestSpeedCalculationWithMissingGPSData tests the smoothing of speed
// data when no GPS point was found between two points.
// This can especially happen when the GPS of the phone is used instead of the devices one
func TestSpeedCalculationWithMissingGPSData(t *testing.T) {
	input := models.Workout{
		SpeedAv: 211, // 17 km/h
		WorkoutDetails: []models.WorkoutDetails{
			{
				Duration: 0,
				Distance: 0,
				Speed:    130, // 27 km/h
			},
			{
				Duration: 6,
				Distance: 35,
				Speed:    171, // 21 km/h
			},
			{
				Duration: 12,
				Distance: 35,
				Speed:    0, // Missing GPS data
			},
			{
				Duration: 18,
				Distance: 35,
				Speed:    0, // Missing GPS data
			},
			{
				Duration: 24,
				Distance: 160,
				Speed:    48, // 75 km/h => Extreme => Replace with avg
			},
			{
				Duration: 30,
				Distance: 180,
				Speed:    300,
			},
			{
				Duration: 36,
				Distance: 200,
				Speed:    300,
			},
			// Negative test: do not modify this point
			{
				Duration: 42,
				Distance: 200,
				Speed:    0,
			},
			{
				Duration: 48,
				Distance: 200,
				Speed:    0,
			},
			{
				Duration: 54,
				Distance: 220,
				Speed:    300,
			},
		},
	}

	processor := NewPostProcessor(PostProcessingOptions{})
	processor.PostProcess(&input)

	expected := models.Workout{
		WorkoutDetails: []models.WorkoutDetails{
			{Speed: 130},
			{Speed: 171},
			{Speed: 144}, // Avg
			{Speed: 144}, // Avg
			{Speed: 144}, // Avg
			{Speed: 300},
			{Speed: 300},
			{Speed: 0},
			{Speed: 0},
			{Speed: 300},
		},
	}

	assert.Equal(t, len(expected.WorkoutDetails), len(input.WorkoutDetails))
	for idx, exp := range expected.WorkoutDetails {
		got := input.WorkoutDetails[idx]
		assert.Equal(t, exp.Speed, got.Speed, fmt.Sprintf("Point #%d at %d seconds", idx, got.Duration))
	}
}
