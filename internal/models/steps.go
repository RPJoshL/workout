package models

import (
	"time"
)

type Steps struct {
	// Start date of the step count
	Start time.Time `json:"start" dbColumn:"Column:start"`
	// End date of the step count
	End time.Time `json:"end" dbColumn:"Column:end"`
	// ID of the user to which the steps belong to
	UserId int `json:"userId" dbColumn:"Column:user_id,ForeignKey:workout.user.id"`
	// The number of steps that were tracked between start and end
	Count       int `json:"count" dbColumn:"Column:count"`
	DbMetadata_ any `json:"-" dbMetadata:"Schema:workout,Table:steps"`
}

// Steps
const (
	Steps_Start  string = "Start|workout.steps.start"
	Steps_End    string = "End|workout.steps.end"
	Steps_UserId string = "UserId|workout.steps.user_id"
	Steps_Count  string = "Count|workout.steps.count"
)
