package workout

import (
	"net/http"
	"strconv"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/cities"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/create"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/details"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/overview"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
)

type Api struct {
	router.ApiRequest

	Overview *overview.Api
	Details  *details.Api
	Create   *create.Api
	City     *cities.Api
}

func GetRoutes(db *database.DatabaseUtils) *router.Router {

	// Initialize types
	shared.InitializeTypes(db)

	api := &Api{
		Overview: &overview.Api{},
		City:     &cities.Api{},
	}

	api.Details = &details.Api{
		Root: api.Overview,
	}
	api.Create = &create.Api{
		Root:    api.Overview,
		Details: api.Details,
	}

	routes := router.Routes{

		// Pages
		router.NewRoute(
			"DeleteWorkout",
			"DELETE",
			"/{id}",
			api.DeleteWorkout,
			router.Options{},
		),
	}

	rout := &router.Router{
		Dependency: api,
		Routes:     routes,
	}

	// Add (sub) routers
	rout.AddRouter(api.Create.GetRouter())
	rout.AddRouter(api.Details.GetRouter())
	rout.AddRouter(api.Overview.GetRouter())
	rout.AddRouter(api.City.GetRouter())

	return rout
}

func (api *Api) DeleteWorkout(w http.ResponseWriter, r *http.Request) {

	// Get ID of workout to display
	workoutId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("#generic.numericError").Sprintf("id", r.PathValue("id")).Write(w, r)
		return
	}

	// Delete workout
	if err := api.Delete(workoutId); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		response.WriteText("Workout deleted", 200, w)
	}
}
