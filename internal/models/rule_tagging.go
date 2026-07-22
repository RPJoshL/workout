package models

import (
	"github.com/guregu/null/v5"
)

type RuleTagging struct {
	// Unique ID of this tagging rule
	Id int `json:"id" dbColumn:"Column:id,AutoIncrement,PrimaryKey"`
	// ID to which this rule belongs to
	UserId int `json:"userId" dbColumn:"Column:user_id,ForeignKey:workout.user.id"`
	// Tag to apply when the rule does match
	TagId int `json:"tagId" dbColumn:"Column:tag_id,ForeignKey:workout.tag.id"`
	// Unique name of this rule for the user
	Name string `json:"name" dbColumn:"Column:name"`
	// Area in which the start point must be located
	StartLocation *AreaCircle `json:"startLocation" dbColumn:"Column:start_location,ForeignKey:workout.area_circle.id,DefaultValue"`
	// Area in which the end point must be located
	EndLocation *AreaCircle `json:"endLocation" dbColumn:"Column:end_location,ForeignKey:workout.area_circle.id,DefaultValue"`
	// Minimum duration in seconds
	DurationMin null.Int64 `json:"durationMin" dbColumn:"Column:duration_min,DefaultValue"`
	// Maximum duration in seconds
	DurationMax null.Int64 `json:"durationMax" dbColumn:"Column:duration_max,DefaultValue"`
	// Downsample the workout points to 30 seconds
	Downsample30 int `json:"downsample30" dbColumn:"Column:downsample_30,DefaultValue"`
	DbMetadata_  any `json:"-" dbMetadata:"Schema:workout,Table:rule_tagging"`
}

// RuleTagging
const (
	RuleTagging_Id            string = "Id|workout.rule_tagging.id"
	RuleTagging_UserId        string = "UserId|workout.rule_tagging.user_id"
	RuleTagging_TagId         string = "TagId|workout.rule_tagging.tag_id"
	RuleTagging_Name          string = "Name|workout.rule_tagging.name"
	RuleTagging_StartLocation string = "StartLocation|workout.rule_tagging.start_location"
	RuleTagging_EndLocation   string = "EndLocation|workout.rule_tagging.end_location"
	RuleTagging_DurationMin   string = "DurationMin|workout.rule_tagging.duration_min"
	RuleTagging_DurationMax   string = "DurationMax|workout.rule_tagging.duration_max"
	RuleTagging_Downsample30  string = "Downsample30|workout.rule_tagging.downsample_30"
)
