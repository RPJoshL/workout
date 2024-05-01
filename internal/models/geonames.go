package models

import (
	"database/sql"
	"git.rpjosh.de/RPJosh/go-ddl-parser"
)

type Geonames struct {
	Geonameid      int            `json:"geonameid" dbColumn:"Column:geonameid,PrimaryKey"`
	Name           string         `json:"name" dbColumn:"Column:name"`
	Alternatenames sql.NullString `json:"alternatenames" dbColumn:"Column:alternatenames,DefaultValue"`
	Location       ddl.Location   `json:"location" dbColumn:"Column:location"`
	Country        string         `json:"country" dbColumn:"Column:country"`
	Population     int            `json:"population" dbColumn:"Column:population"`
	DbMetadata_    any            `json:"-" dbMetadata:"Schema:workout,Table:geonames"`
}

// Geonames
const (
	Geonames_Geonameid      string = "Geonameid|workout.geonames.geonameid"
	Geonames_Name           string = "Name|workout.geonames.name"
	Geonames_Alternatenames string = "Alternatenames|workout.geonames.alternatenames"
	Geonames_Location       string = "Location|workout.geonames.location"
	Geonames_Country        string = "Country|workout.geonames.country"
	Geonames_Population     string = "Population|workout.geonames.population"
)
