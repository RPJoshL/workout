package models

import (
	"git.rpjosh.de/RPJosh/go-ddl-parser"
)

type AreaCircle struct {
	// Unique ID of this location area
	Id int `json:"id" dbColumn:"Column:id,AutoIncrement,PrimaryKey"`
	// Center of this circle area
	Center ddl.Location `json:"center" dbColumn:"Column:center"`
	// Radius in meters of the circle area
	Radius      int `json:"radius" dbColumn:"Column:radius"`
	DbMetadata_ any `json:"-" dbMetadata:"Schema:workout,Table:area_circle"`
}

// AreaCircle
const (
	AreaCircle_Id     string = "Id|workout.area_circle.id"
	AreaCircle_Center string = "Center|workout.area_circle.center"
	AreaCircle_Radius string = "Radius|workout.area_circle.radius"
)
