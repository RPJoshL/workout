package leaflet

import (
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
)

// Map renders an interactive map with the JavaScript
// library "Leaflet"
type Map struct {
	T *translator.Translator
}

// Options contains generic options to configure
// the behaviour of the map
type Options struct {

	// Internal ID of the div that display the leaflet map
	id string

	// Points to display in the map as a SINGLE connected line
	Line []Point

	// Lines to display on the map
	Lines []Line
}

// Point is a single point in the map
type Point struct {
	Latitude  float64
	Longitude float64

	// Distance in meters since the beginning of the line
	Distance int

	// Optional heartrate to color the lines for
	Heartrate int

	// Content of a hoovered popup as a raw HTML value
	TooltipContent string
}

// Line is a wrapper around []Point with additional details
// for a single line
type Line struct {

	// Content of a popup when hovering over the line as a raw HTML value
	TooltipContent string

	// Points that descibes this line.
	// Only the "Latitude" and "Longitude" fields are used
	// from the Javascript Library
	Points []Point
}

func (o *Options) getId() string {
	if o.id == "" {
		o.id, _ = utils.GenerateRandomString(12)
		o.id = "o" + o.id
	}

	return o.id
}
