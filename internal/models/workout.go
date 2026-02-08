package models

import (
	"fmt"
	"strings"
	"time"

	"git.rpjosh.de/RPJosh/go-ddl-parser"
	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"github.com/guregu/null/v5"
)

const (
	GENDER_MALE = iota
	GENDER_FEMALE
)

const (
	TYPE_UNKNOWN = iota
	TYPE_HIKING
	TYPE_RUNNING
	TYPE_SURFEN
	TYPE_SAILING
	TYPE_SNOWBOARDING
	TYPE_SWIMMING
	TYPE_CYCLING
	TYPE_SKATEBOARDING
	TYPE_VOLLEYBALL
	TYPE_PUMP_FOILING
	TYPE_STRENGTH_TRAINING
	TYPE_ICE_SKATING
)

// TypeNameMap is a map that contains different acitivty names in
// various languages and contexts.
// It's indexed by a type value.
//
// Important: the first value is ALWAYS the string key of that
// workout type
var TypeNameMap = map[int][]string{
	TYPE_HIKING:            {"walking", "gehen"},
	TYPE_RUNNING:           {"running", "joggen", "laufen"},
	TYPE_SURFEN:            {"surf", "windsurfen"},
	TYPE_SAILING:           {"sailing", "segeln"},
	TYPE_SNOWBOARDING:      {"snowboarding", "snowboarden"},
	TYPE_SWIMMING:          {"swimming", "schwimmen"},
	TYPE_CYCLING:           {"cycling", "radfahren", "biking"},
	TYPE_SKATEBOARDING:     {"skateboarding", "skaten"},
	TYPE_VOLLEYBALL:        {"volleyball", "beachvolleyball"},
	TYPE_PUMP_FOILING:      {"pumping", "pump foiling", "foiling"},
	TYPE_STRENGTH_TRAINING: {"strength trainings", "weight lifing", "lifting", "krafttraining", "fitness", "bodybuilding"},
	TYPE_ICE_SKATING:       {"ice skating", "eislaufen", "schlittschuhlaufen", "schlittschuh"},
}

const (
	TYPE_CATEGORY_SNOW    = "SNOW"
	TYPE_CATEGORY_WATER   = "WATER"
	TYPE_CATEGORY_BALL    = "BALL"
	TYPE_CATEGORY_WALKING = "WALKING"
	TYPE_CATEGORY_CYCLING = "CYCLING"
	// All other outdoor activities (skateboard, inliner, fishing, ...)
	TYPE_CATEGORY_OUTDOOR = "OUTDOOR"
)

type SamplingLevel uint8

const (
	// 6 seconds between each point
	SamplingLevelDefault SamplingLevel = iota
	// 3 seconds between each point
	SamplingLevelDetailed
	// 30 seconds with Ramer–Douglas–Peucker algorithm for GPS track
	SamplingLevelDownsampled
)

type Workout struct {
	// Unique ID of the workout
	Id int `json:"id" dbColumn:"Column:id,AutoIncrement,PrimaryKey"`
	// Name that describes this workout
	Name string `json:"name" dbColumn:"Column:name"`
	// ID of the user the workout belongs to
	UserId int `json:"userId" dbColumn:"Column:user_id,ForeignKey:workout.user.id"`
	// Workout type or categorie
	TypeId int `json:"typeId" dbColumn:"Column:type_id,ForeignKey:workout.workout_type.id"`
	// Time and date the workout was started
	Start time.Time `json:"start" dbColumn:"Column:start"`
	// Time and date the workout was completed
	End time.Time `json:"end" dbColumn:"Column:end"`
	// 2 letter country code where the workout was started
	Country string `json:"country" dbColumn:"Column:country"`
	// Name of the city where the workout was started
	City string `json:"city" dbColumn:"Column:city"`
	// Unique ID for the city in the geonames database where the workout was started
	CityId int `json:"cityId" dbColumn:"Column:city_id"`
	// Latitude of the city
	CityLocation ddl.Location `json:"cityLocation" dbColumn:"Column:city_location"`
	// Duration in seconds the workout lasted without any pauses
	Duration int `json:"duration" dbColumn:"Column:duration"`
	// Number of calories that were burned during the workout "duration"
	Calories int `json:"calories" dbColumn:"Column:calories"`
	// Number of calories that were by default burned during the workout "duration"
	CaloriesDefault int `json:"caloriesDefault" dbColumn:"Column:calories_default"`
	// Distance in meters traveled during the workout
	Distance int `json:"distance" dbColumn:"Column:distance"`
	// Average traveling speed in sec/km
	SpeedAv int `json:"speedAv" dbColumn:"Column:speed_av"`
	// Attitude meters (up) made during the workout
	ElevationUp int `json:"elevationUp" dbColumn:"Column:elevation_up"`
	// Attitude meters (down) made during the workout
	ElevationDown int `json:"elevationDown" dbColumn:"Column:elevation_down"`
	// Average heart rate during the workout
	HeartRateAv null.Int64 `json:"heartRateAv" dbColumn:"Column:heart_rate_av,DefaultValue"`
	// Maximum heart rate during the workout
	HeartRateMax null.Int64 `json:"heartRateMax" dbColumn:"Column:heart_rate_max,DefaultValue"`
	// Text describing this workout in Markdown format
	Note null.String `json:"note" dbColumn:"Column:note,DefaultValue"`
	Pai  int         `json:"pai" dbColumn:"Column:pai,DefaultValue"`
	// Number of steps that were made during the entire workout
	Steps null.Int64 `json:"steps" dbColumn:"Column:steps,DefaultValue"`
	// Level of downsampling that was applied to the workout details
	SamplingLevel  int              `json:"samplingLevel" dbColumn:"Column:sampling_level,DefaultValue"`
	WorkoutDetails []WorkoutDetails `dbColumn:"PointedForeignKey:workout.workout_details.workout_id"`
	WorkoutTags    []WorkoutTags    `dbColumn:"PointedForeignKey:workout.workout_tags.workout_id"`
	DbMetadata_    any              `json:"-" dbMetadata:"Schema:workout,Table:workout"`
}

// Workout
const (
	Workout_Id              string = "Id|workout.workout.id"
	Workout_Name            string = "Name|workout.workout.name"
	Workout_UserId          string = "UserId|workout.workout.user_id"
	Workout_TypeId          string = "TypeId|workout.workout.type_id"
	Workout_Start           string = "Start|workout.workout.start"
	Workout_End             string = "End|workout.workout.end"
	Workout_Country         string = "Country|workout.workout.country"
	Workout_City            string = "City|workout.workout.city"
	Workout_CityId          string = "CityId|workout.workout.city_id"
	Workout_CityLocation    string = "CityLocation|workout.workout.city_location"
	Workout_Duration        string = "Duration|workout.workout.duration"
	Workout_Calories        string = "Calories|workout.workout.calories"
	Workout_CaloriesDefault string = "CaloriesDefault|workout.workout.calories_default"
	Workout_Distance        string = "Distance|workout.workout.distance"
	Workout_SpeedAv         string = "SpeedAv|workout.workout.speed_av"
	Workout_ElevationUp     string = "ElevationUp|workout.workout.elevation_up"
	Workout_ElevationDown   string = "ElevationDown|workout.workout.elevation_down"
	Workout_HeartRateAv     string = "HeartRateAv|workout.workout.heart_rate_av"
	Workout_HeartRateMax    string = "HeartRateMax|workout.workout.heart_rate_max"
	Workout_Note            string = "Note|workout.workout.note"
	Workout_Pai             string = "Pai|workout.workout.pai"
	Workout_Steps           string = "Steps|workout.workout.steps"
	Workout_SamplingLevel   string = "SamplingLevel|workout.workout.sampling_level"
	Workout_WorkoutDetails  string = "WorkoutDetails|#workout.workout.WorkoutDetails"
	Workout_WorkoutTags     string = "WorkoutTags|#workout.workout.WorkoutTags"
)

type WorkoutDetails struct {
	// Unique ID of the workout details
	Id int `json:"id" dbColumn:"Column:id,AutoIncrement,PrimaryKey"`
	// Workout reference
	WorkoutId int `json:"workoutId" dbColumn:"Column:workout_id,ForeignKey:workout.workout.id"`
	// There are two different types of workout details stored:
	// 0 = detailed and all workout points | 1 = downsampled points for an overview table
	Type int `json:"type" dbColumn:"Column:type"`
	// Duration (without pauses) since the beginning of the workout in seconds
	Duration int `json:"duration" dbColumn:"Column:duration"`
	// Date and time of this point
	Time time.Time `json:"time" dbColumn:"Column:time,DefaultValue"`
	// Distance in meters traveled for this point from the beginning of the workout (without pauses)
	Distance int `json:"distance" dbColumn:"Column:distance"`
	// Longitude of the data point
	Longitude float64 `json:"longitude" dbColumn:"Column:longitude"`
	// Latitude of the data point
	Latitude float64 `json:"latitude" dbColumn:"Column:latitude"`
	// Elevation height of the data point. This can be 0 if elevation is not supported by the tracker
	Elevation int `json:"elevation" dbColumn:"Column:elevation"`
	// Cummolated traveling speed in sec/km
	Speed int `json:"speed" dbColumn:"Column:speed"`
	// Current heart rate
	HeartRate null.Int64 `json:"heartRate" dbColumn:"Column:heart_rate,DefaultValue"`
	// Number of total steps made since the beginning of the workout
	StepCount null.Int64 `json:"stepCount" dbColumn:"Column:step_count,DefaultValue"`
	// Part / track index when merging multiple workouts into a single one
	Part        int `json:"part" dbColumn:"Column:part,DefaultValue"`
	DbMetadata_ any `json:"-" dbMetadata:"Schema:workout,Table:workout_details"`
}

// WorkoutDetails
const (
	WorkoutDetails_Id        string = "Id|workout.workout_details.id"
	WorkoutDetails_WorkoutId string = "WorkoutId|workout.workout_details.workout_id"
	WorkoutDetails_Type      string = "Type|workout.workout_details.type"
	WorkoutDetails_Duration  string = "Duration|workout.workout_details.duration"
	WorkoutDetails_Time      string = "Time|workout.workout_details.time"
	WorkoutDetails_Distance  string = "Distance|workout.workout_details.distance"
	WorkoutDetails_Longitude string = "Longitude|workout.workout_details.longitude"
	WorkoutDetails_Latitude  string = "Latitude|workout.workout_details.latitude"
	WorkoutDetails_Elevation string = "Elevation|workout.workout_details.elevation"
	WorkoutDetails_Speed     string = "Speed|workout.workout_details.speed"
	WorkoutDetails_HeartRate string = "HeartRate|workout.workout_details.heart_rate"
	WorkoutDetails_StepCount string = "StepCount|workout.workout_details.step_count"
	WorkoutDetails_Part      string = "Part|workout.workout_details.part"
)

type WorkoutTags struct {
	// Reference to workout
	WorkoutId int `json:"workoutId" dbColumn:"Column:workout_id,PrimaryKey,ForeignKey:workout.workout.id"`
	// Reference to assigned tag
	TagId       *Tag `json:"tagId" dbColumn:"Column:tag_id,PrimaryKey,ForeignKey:workout.tag.id"`
	DbMetadata_ any  `json:"-" dbMetadata:"Schema:workout,Table:workout_tags"`
}

// WorkoutTags
const (
	WorkoutTags_WorkoutId string = "WorkoutId|workout.workout_tags.workout_id"
	WorkoutTags_TagId     string = "TagId|workout.workout_tags.tag_id"
)

type WorkoutType struct {
	// Unique ID of this workout type
	Id int `json:"id" dbColumn:"Column:id,AutoIncrement,PrimaryKey"`
	// Description name of the workout type
	NameDe string `json:"nameDe" dbColumn:"Column:name_de"`
	// Description name of the workout type (EN)
	NameEn string `json:"nameEn" dbColumn:"Column:name_en"`
	// Color code (#f20102) of the tag for the dark mode
	TagDark string `json:"tagDark" dbColumn:"Column:tag_dark"`
	// Color code (#f20102) of the tag for the white mode
	TagWhite string `json:"tagWhite" dbColumn:"Column:tag_white"`
	// Category of the workout type like "SNOW", "WATER", "WALKING"
	Category    string `json:"category" dbColumn:"Column:category,DefaultValue"`
	DbMetadata_ any    `json:"-" dbMetadata:"Schema:workout,Table:workout_type"`
}

// WorkoutType
const (
	WorkoutType_Id       string = "Id|workout.workout_type.id"
	WorkoutType_NameDe   string = "NameDe|workout.workout_type.name_de"
	WorkoutType_NameEn   string = "NameEn|workout.workout_type.name_en"
	WorkoutType_TagDark  string = "TagDark|workout.workout_type.tag_dark"
	WorkoutType_TagWhite string = "TagWhite|workout.workout_type.tag_white"
	WorkoutType_Category string = "Category|workout.workout_type.category"
)

// GetWorkoutTypeByName returns a matching workout type
// by the provided string
func GetWorkoutTypeByName(name string) int {
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)

	// Nothing to compare against
	if name == "" || len(name) < 2 {
		return TYPE_UNKNOWN
	}

	// Try to match whole word
	for key, vals := range TypeNameMap {
		for _, val := range vals {
			if strings.EqualFold(val, name) {
				return key
			}
		}
	}

	// Try to match if contained in name
	for key, vals := range TypeNameMap {
		for _, val := range vals {
			if strings.Contains(strings.ToLower(val), name) {
				return key
			}
		}
	}

	logger.Trace("Did not found a workout type for %q", name)
	return TYPE_UNKNOWN
}

// AvgSpeedInKmPerHour returns the average traveling speed in km/h
func (w *Workout) AvgSpeedInKmPerHour() float64 {
	return 1.0 / (float64(w.SpeedAv) / 3600)
}

// AvgSpeedInKmPerHour returns the average traveling speed in km/h
func (d *WorkoutDetails) AvgSpeedInKmPerHour() float64 {
	rtc := 1.0 / (float64(d.Speed) / 3600)
	if d.Speed == 0 {
		// Don't display inf
		rtc = 0
	}

	return rtc
}

func (w *Workout) GetDuration() string {
	return formatDuration(w.Duration)
}

func (w *Workout) GetDistance() string {
	return fmt.Sprintf("%.2f km", float64(w.Distance)/1000.0)
}

// GetDuration returns a nicely formatted duration to display
// in whe Webapp
func (d *WorkoutDetails) GetDuration() string {
	return formatDuration(d.Duration)
}

// formatDuration formats the provided duration nicely in the
// format "h m s"
func formatDuration(duration int) string {
	// Only display hours for duration > 100 minutes
	if duration >= (100 * 60) {
		return fmt.Sprintf("%dh %02dm %02ds", duration/3600, (duration/60)%60, duration%60)
	} else {
		return fmt.Sprintf("%dm %02ds", duration/60, duration%60)
	}
}

// GetNameForLanguage returns the name of the workout type for the provided
// language
func (t WorkoutType) GetNameForLanguage(language translator.Language) string {
	switch language {
	case translator.English:
		return t.NameEn
	case translator.German:
		return t.NameDe
	}

	return t.NameEn
}
