package overview

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/utils"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/cities"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/details"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/a-h/templ"
)

type Api struct {
	router.ApiRequest

	City    cities.Api
	Details *details.Api

	Shared shared.Shared
}

func (api *Api) GetRouter() *router.Router {
	api.Details = &details.Api{
		Root: api,
	}

	routes := router.Routes{
		router.NewRoute(
			"WorkoutTablePage",
			"GET",
			"/",
			api.GetWorkoutTablePage,
			router.Options{},
		),
		router.NewRoute(
			"WorkoutTableDataElement",
			"GET",
			"/tableData",
			api.GetWorkoutTableData,
			router.Options{},
		),
		router.NewRoute(
			"WorkoutMapOverview",
			"GET",
			"/map",
			api.GetWorkoutOverviewMap,
			router.Options{},
		),
		router.NewRoute(
			"WorkoutTablePageListPopup",
			"GET",
			"/{id}/listPopup",
			api.DetailsListPopup,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

// GetWorkoutTablePage returns the complete overview page with a
// default workout selection
func (api *Api) GetWorkoutTablePage(w http.ResponseWriter, r *http.Request) {
	comp := api.getWorkoutTablePage(w, r)
	if comp != nil {
		api.R().Tmpl.Render(comp, "generic.appName", "generic.appName")
	}
}

// getWorkoutTablePage returns the complete overview page with a
// default workout selection.
//
// Errors are already written to the response
func (api *Api) getWorkoutTablePage(w http.ResponseWriter, r *http.Request) templ.Component {
	// Get data to display
	data, e := api.GetTableData(false, shared.WorkoutFilter{
		DateRange: fmt.Sprintf(
			"%s to %s",
			time.Now().AddDate(0, -3, 0).Format("02.01.2006"),
			time.Now().Format("02.01.2006"),
		),
	})
	if e != nil {
		e.GetErrorStruct().Write(w, r)
		return nil
	}

	return api.MainWithData(data)
}

// GetWorkoutTableData returns the table and list element with workout data
// filtered by the provided request parameters
func (api *Api) GetWorkoutTableData(w http.ResponseWriter, r *http.Request) {

	// Get filter
	filter := shared.WorkoutFilter{}
	if err := api.R().Parser.Parse(&filter); err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	// Get data to display
	data, e := api.GetTableData(false, filter)
	if e != nil {
		e.GetErrorStruct().Write(w, r)
		return
	}

	// Render table elements (with list) directly
	api.R().Tmpl.RenderDirect(api.OverviewData(data))
}
func (api *Api) Main() (templ.Component, string) {
	req, resp := api.R().GetHttpRequest()
	return api.getWorkoutTablePage(resp, req), utils.GetCallerFile()
}

// GetWorkoutOverviewMap displays a workout map with all (filtered)
// workouts
func (api *Api) GetWorkoutOverviewMap(w http.ResponseWriter, r *http.Request) {

	// Get filter
	filter := shared.WorkoutFilter{}
	if err := api.R().Parser.Parse(&filter); err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	// Get data to display
	data, e := api.GetTableData(true, filter)
	if e != nil {
		e.GetErrorStruct().Write(w, r)
		return
	}

	// Render map element directly
	api.R().Tmpl.RenderDirect(api.OverviewMap(data))
}

func (api *Api) DetailsListPopup(w http.ResponseWriter, r *http.Request) {
	// Get ID of workout to display
	workoutId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("#generic.numericError").Sprintf("id", r.PathValue("id")).Write(w, r)
		return
	}

	api.R().Tmpl.RenderDirect(api.listPopup(workoutId))
}
