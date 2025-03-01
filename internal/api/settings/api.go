package settings

import (
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/settings/automation"
	"git.rpjosh.de/RPJosh/workout/internal/api/settings/generic"
	"git.rpjosh.de/RPJosh/workout/internal/api/settings/token"
)

type Api struct {
	router.ApiRequest

	Generic    *generic.Api
	Token      *token.Api
	Automation *automation.Api
}

func GetRoutes() *router.Router {

	api := &Api{
		Generic:    &generic.Api{},
		Token:      &token.Api{},
		Automation: &automation.Api{},
	}

	// No direct routes in settings
	routes := router.Routes{}

	rout := &router.Router{
		Dependency: api,
		Routes:     routes,
	}

	// Add (sub) routers
	rout.AddRouter(api.Generic.GetRouter())
	rout.AddRouter(api.Token.GetRouter())
	rout.AddRouter(api.Automation.GetRouter())

	return rout
}
