package chart

import "git.rpjosh.de/RPJosh/workout/pkg/utils"

// Chart provides various chart diagramms to display
// time series data
type Chart struct{}

type Options struct {

	// Internal ID of the div to render the chart in
	id string

	// Weather to use a dark theme for charts
	DarkTheme bool

	// Name of the exported JavaScript method to render the chart with
	RenderMethod string

	// Data that is passed to the chart
	Data any
}

func (o *Options) getId() string {
	if o.id == "" {
		o.id, _ = utils.GenerateRandomString(12)
		o.id = "o" + o.id
	}

	return o.id
}
