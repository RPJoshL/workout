package models

import (
	"github.com/RPJoshL/go-ddl-parser"
	"github.com/guregu/null/v5"
	"strings"
)

type Geonames struct {
	Geonameid      int          `json:"geonameid" dbColumn:"Column:geonameid,PrimaryKey"`
	Name           string       `json:"name" dbColumn:"Column:name"`
	Alternatenames null.String  `json:"alternatenames" dbColumn:"Column:alternatenames,DefaultValue"`
	Location       ddl.Location `json:"location" dbColumn:"Column:location"`
	Country        string       `json:"country" dbColumn:"Column:country"`
	Population     int          `json:"population" dbColumn:"Column:population"`
	Adm1           null.String  `json:"adm1" dbColumn:"Column:adm1,DefaultValue"`
	Adm2           null.String  `json:"adm2" dbColumn:"Column:adm2,DefaultValue"`
	Adm3           null.String  `json:"adm3" dbColumn:"Column:adm3,DefaultValue"`
	Adm4           null.String  `json:"adm4" dbColumn:"Column:adm4,DefaultValue"`
	DbMetadata_    any          `json:"-" dbMetadata:"Schema:workout,Table:geonames"`
}

// Geonames
const (
	Geonames_Geonameid      string = "Geonameid|workout.geonames.geonameid"
	Geonames_Name           string = "Name|workout.geonames.name"
	Geonames_Alternatenames string = "Alternatenames|workout.geonames.alternatenames"
	Geonames_Location       string = "Location|workout.geonames.location"
	Geonames_Country        string = "Country|workout.geonames.country"
	Geonames_Population     string = "Population|workout.geonames.population"
	Geonames_Adm1           string = "Adm1|workout.geonames.adm1"
	Geonames_Adm2           string = "Adm2|workout.geonames.adm2"
	Geonames_Adm3           string = "Adm3|workout.geonames.adm3"
	Geonames_Adm4           string = "Adm4|workout.geonames.adm4"
)

type GeonamesAdm struct {
	Geonameid      int         `json:"geonameid" dbColumn:"Column:geonameid,PrimaryKey"`
	Typ            string      `json:"typ" dbColumn:"Column:typ"`
	Value          string      `json:"value" dbColumn:"Column:value"`
	Name           string      `json:"name" dbColumn:"Column:name"`
	Alternatenames null.String `json:"alternatenames" dbColumn:"Column:alternatenames,DefaultValue"`
	Adm0           string      `json:"adm0" dbColumn:"Column:adm0"`
	Adm1           string      `json:"adm1" dbColumn:"Column:adm1"`
	Adm2           null.String `json:"adm2" dbColumn:"Column:adm2,DefaultValue"`
	Adm3           null.String `json:"adm3" dbColumn:"Column:adm3,DefaultValue"`
	Root           null.Int64  `json:"root" dbColumn:"Column:root,DefaultValue"`
	DbMetadata_    any         `json:"-" dbMetadata:"Schema:workout,Table:geonames_adm"`
}

// GeonamesAdm
const (
	GeonamesAdm_Geonameid      string = "Geonameid|workout.geonames_adm.geonameid"
	GeonamesAdm_Typ            string = "Typ|workout.geonames_adm.typ"
	GeonamesAdm_Value          string = "Value|workout.geonames_adm.value"
	GeonamesAdm_Name           string = "Name|workout.geonames_adm.name"
	GeonamesAdm_Alternatenames string = "Alternatenames|workout.geonames_adm.alternatenames"
	GeonamesAdm_Adm0           string = "Adm0|workout.geonames_adm.adm0"
	GeonamesAdm_Adm1           string = "Adm1|workout.geonames_adm.adm1"
	GeonamesAdm_Adm2           string = "Adm2|workout.geonames_adm.adm2"
	GeonamesAdm_Adm3           string = "Adm3|workout.geonames_adm.adm3"
	GeonamesAdm_Root           string = "Root|workout.geonames_adm.root"
)

type VGeonamesAll struct {
	Geonameid      int          `json:"geonameid" dbColumn:"Column:geonameid"`
	Name           string       `json:"name" dbColumn:"Column:name"`
	Alternatenames null.String  `json:"alternatenames" dbColumn:"Column:alternatenames,DefaultValue"`
	Location       ddl.Location `json:"location" dbColumn:"Column:location"`
	Country        string       `json:"country" dbColumn:"Column:country"`
	Population     int          `json:"population" dbColumn:"Column:population"`
	Adm1           null.String  `json:"adm1" dbColumn:"Column:adm1,DefaultValue"`
	Adm2           null.String  `json:"adm2" dbColumn:"Column:adm2,DefaultValue"`
	Adm3           null.String  `json:"adm3" dbColumn:"Column:adm3,DefaultValue"`
	Adm4           null.String  `json:"adm4" dbColumn:"Column:adm4,DefaultValue"`
	DisplayName    null.String  `json:"displayName" dbColumn:"Column:display_name,DefaultValue"`
	Adm3Name       null.String  `json:"adm3Name" dbColumn:"Column:adm3_name"`
	Adm2Name       null.String  `json:"adm2Name" dbColumn:"Column:adm2_name"`
	Adm1Name       null.String  `json:"adm1Name" dbColumn:"Column:adm1_name"`
	DbMetadata_    any          `json:"-" dbMetadata:"Schema:workout,Table:v_geonames_all"`
}

// VGeonamesAll
const (
	VGeonamesAll_Geonameid      string = "Geonameid|workout.v_geonames_all.geonameid"
	VGeonamesAll_Name           string = "Name|workout.v_geonames_all.name"
	VGeonamesAll_Alternatenames string = "Alternatenames|workout.v_geonames_all.alternatenames"
	VGeonamesAll_Location       string = "Location|workout.v_geonames_all.location"
	VGeonamesAll_Country        string = "Country|workout.v_geonames_all.country"
	VGeonamesAll_Population     string = "Population|workout.v_geonames_all.population"
	VGeonamesAll_Adm1           string = "Adm1|workout.v_geonames_all.adm1"
	VGeonamesAll_Adm2           string = "Adm2|workout.v_geonames_all.adm2"
	VGeonamesAll_Adm3           string = "Adm3|workout.v_geonames_all.adm3"
	VGeonamesAll_Adm4           string = "Adm4|workout.v_geonames_all.adm4"
	VGeonamesAll_DisplayName    string = "DisplayName|workout.v_geonames_all.display_name"
	VGeonamesAll_Adm3Name       string = "Adm3Name|workout.v_geonames_all.adm3_name"
	VGeonamesAll_Adm2Name       string = "Adm2Name|workout.v_geonames_all.adm2_name"
	VGeonamesAll_Adm1Name       string = "Adm1Name|workout.v_geonames_all.adm1_name"
)

// GetFullName returns all admin codes (from 1 - 3) separated by
// a comma and enclosed with brackets
func (v *VGeonamesAll) GetFullName() string {
	rtc := ""
	if v.Adm3Name.Valid && v.Adm3Name.String != v.DisplayName.String {
		rtc += replaceGenericDetailsInGeonameName(v.Adm3Name.String)
	}
	if v.Adm2Name.Valid {
		if rtc != "" {
			rtc += ", "
		}
		rtc += replaceGenericDetailsInGeonameName(v.Adm2Name.String)
	}
	if v.Adm1Name.Valid {
		if rtc != "" {
			rtc += ", "
		}
		rtc += replaceGenericDetailsInGeonameName(v.Adm1Name.String)
	}

	if rtc != "" {
		return "(" + rtc + ")"
	} else {
		return rtc
	}
}

// replaceGenericDetails replaces some default expressions
// from the administrative name like 'Landkreis', 'Politischer Berzirk', ...
func replaceGenericDetailsInGeonameName(name string) string {
	prefixe := []string{
		"Landkreis ",
		"Politischer Bezirk ",
	}
	for _, p := range prefixe {
		if after, ok := strings.CutPrefix(name, p); ok {
			return after
		}
	}

	return name
}
