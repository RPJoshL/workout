package models

import (
	"github.com/guregu/null/v5"
)

type PaiDaily struct {
	Id           int        `json:"id" dbColumn:"Column:id"`
	StepsTotal   float64    `json:"stepsTotal" dbColumn:"Column:steps_total,DefaultValue"`
	StepsWorkout float64    `json:"stepsWorkout" dbColumn:"Column:steps_workout,DefaultValue"`
	WorkoutPai   float64    `json:"workoutPai" dbColumn:"Column:workout_pai,DefaultValue"`
	StepsPai     int        `json:"stepsPai" dbColumn:"Column:steps_pai,DefaultValue"`
	UserId       null.Int64 `json:"userId" dbColumn:"Column:user_id,DefaultValue"`
	DbMetadata_  any        `json:"-" dbMetadata:"Schema:workout,Table:pai_daily"`
}

// PaiDaily
const (
	PaiDaily_Id           string = "Id|workout.pai_daily.id"
	PaiDaily_StepsTotal   string = "StepsTotal|workout.pai_daily.steps_total"
	PaiDaily_StepsWorkout string = "StepsWorkout|workout.pai_daily.steps_workout"
	PaiDaily_WorkoutPai   string = "WorkoutPai|workout.pai_daily.workout_pai"
	PaiDaily_StepsPai     string = "StepsPai|workout.pai_daily.steps_pai"
	PaiDaily_UserId       string = "UserId|workout.pai_daily.user_id"
)
