package cities

import (
	"strings"

	selectbox "git.rpjosh.de/RPJosh/workout/internal/api/components/select"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/a-h/templ"
)

var (
	ErrCityToShort = errors.NewError("#workout.cityNameToShort", 400)
)

// GetCityOptions returns a list of selectable cities from the GeoDB
// database filtered by the provided user input
func (a *Api) GetCityOptions(input string) (templ.Component, errors.Error) {
	if len(input) < 4 {
		return nil, ErrCityToShort
	}
	input = strings.ToUpper(input)

	// Search for city
	cities := []models.VGeonamesAll{}
	sel := a.R().Db.Struct.QuerySlice(&cities)

	// Add where statements
	sel.Where().Custom(`
		UPPER(name) LIKE CONCAT('%', ?, '%') OR UPPER(alternatenames) LIKE CONCAT('%', ?, '%')`,
		input, input,
	).Add()

	// Add order by "best match"
	sel.CustomOrderBy(`
		CASE
			WHEN name         LIKE CONCAT(?, '%') THEN 1
			WHEN display_name LIKE CONCAT(?, '%') THEN 2
			WHEN name         LIKE CONCAT('%', ?, '%') THEN 3
			WHEN display_name LIKE CONCAT('%', ?, '%') THEN 4
			ELSE 5
		END
	`, input, input, input, input)

	if err := sel.Run(); err != nil {
		return nil, err.GetResponse().Log("Failed to query geonames", err, a)
	}

	// Get options
	options := shared.GetGeonamesOptions(cities)

	// Return component
	return a.R().Comp.Select.GetItems("city-selector", options, a.GetCitySearchSettings()), nil
}

// getCitySearchSettings returns the settings to apply for the
// select with the available cities
func (api *Api) GetCitySearchSettings() *selectbox.Settings {
	return &selectbox.Settings{
		Name: "city",
		Remote: selectbox.RemoteFetch{
			Enabled: true,
			Path:    "/workout/city",
			OnEnter: true,
		},
	}
}
