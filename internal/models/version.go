package models

import (
	"time"
)

type Version struct {
	Release     string    `json:"release" dbColumn:"Column:release,PrimaryKey"`
	UpdateTime  time.Time `json:"updateTime" dbColumn:"Column:update_time,DefaultValue"`
	DbMetadata_ any       `json:"-" dbMetadata:"Schema:workout,Table:version"`
}

// Version
const (
	Version_Release    string = "Release|workout.version.release"
	Version_UpdateTime string = "UpdateTime|workout.version.update_time"
)
