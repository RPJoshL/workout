package settings

import (
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/settings/generic"
)

type Api struct {
	router.ApiRequest

	Generic *generic.Api
}

func GetRoutes() *router.Router {

	api := &Api{
		Generic: &generic.Api{},
	}

	// No direct routes in settings
	routes := router.Routes{}

	rout := &router.Router{
		Dependency: api,
		Routes:     routes,
	}

	// Add (sub) routers
	rout.AddRouter(api.Generic.GetRouter())

	return rout
}
