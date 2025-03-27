package shared

import "git.rpjosh.de/RPJosh/workout/pkg/errors"

var (
	ErrOperator = errors.NewError("Invalid comparison operator provided: %q", 400)
)

// WorkoutFilter contains filter conditions for fetching workouts
type WorkoutFilter struct {
	Activities []int `query:"types"`
	Tags       []int `query:"tags"`

	Km         int    `query:"km"`
	KmOperator string `query:"kmOperator"`

	Duration         int    `query:"duration"`
	DurationOperator string `query:"durationOperator"`

	Radius         int    `query:"radius"`
	RadiusOperator string `query:"radiusOperator"`

	City           int    `query:"city"`
	DateRange      string `query:"dateRange"`
	ShowHiddenTags bool   `query:"showHiddenTags"`
}

// ValidateFilterOperator checks if the provided filter operator
// is valid. It's needed to avoid SQL injections
func (f WorkoutFilter) ValidateFilterOperator() errors.Error {
	valsTocheck := []string{f.DurationOperator, f.KmOperator, f.RadiusOperator}
	for _, v := range valsTocheck {
		if v != "" && v != "=" && v != ">" && v != ">=" && v != "<" && v != "<=" && v != "<>" {
			return ErrOperator.Sprintf(v)
		}
	}

	return nil
}
