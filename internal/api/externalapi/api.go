package externalapi

import (
	"net/http"
	"strconv"

	"git.rpjosh.de/RPJosh/workout/internal/api/components/form"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/externalapi"
	"git.rpjosh.de/RPJosh/workout/internal/externalapi/pumpfoilorg"
	"git.rpjosh.de/RPJosh/workout/internal/externalapi/strava"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
	"github.com/a-h/templ"
)

type Api struct {
	router.ApiRequest
	apis      []externalapi.API
	apiLookup map[models.ExternalAPIType]externalapi.API
}

func GetExternalAPIs() (apis []externalapi.API, apiLookup map[models.ExternalAPIType]externalapi.API) {
	apis = []externalapi.API{
		&pumpfoilorg.Api{},
		&strava.Api{},
	}

	apiLookup = make(map[models.ExternalAPIType]externalapi.API, len(apis))
	for _, api := range apis {
		apiLookup[api.Config().Type] = api
	}

	return apis, apiLookup
}

func NewAPI() *Api {
	apis, apiLookup := GetExternalAPIs()

	return &Api{
		apis:      apis,
		apiLookup: apiLookup,
	}
}

func (a *Api) GetSettingRoutes() *router.Router {
	routes := router.Routes{
		router.NewRoute(
			"Overview",
			"GET",
			"/externalApi/",
			a.OverviewPage,
			router.Options{},
		),
		router.NewRoute(
			"GetSupportedTypes",
			"GET",
			"/externalApi",
			a.GetApiTypes,
			router.Options{
				IsApiEndpoint: true,
			},
		),
		router.NewRoute(
			"ServiceTab",
			"GET",
			"/externalApi/{type}",
			a.ServiceTab,
			router.Options{},
		),
		router.NewRoute(
			"Authentication step",
			"POST",
			"/externalApi/{type}/auth/{step}",
			a.Authenticate,
			router.Options{},
		),
		router.NewRoute(
			"Delete authentication",
			"DELETE",
			"/externalApi/{type}",
			a.DeleteSession,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: a,
		Routes:     routes,
	}
}

func GetUploadRoutes() *router.Router {
	api := NewAPI()

	routes := router.Routes{
		router.NewRoute(
			"Upload Workout",
			"POST",
			"/upload/{id}/{type}",
			api.UploadWorkout,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (a *Api) OverviewPage(w http.ResponseWriter, r *http.Request) {
	a.renderOverview(w, r, a.apis[0].Config().Type)
}

type externalAPI struct {
	Type           models.ExternalAPIType `json:"type"`
	Name           string                 `json:"name"`
	SupportedTypes []int                  `json:"supportedTypes"`
}

func (a *Api) GetApiTypes(w http.ResponseWriter, r *http.Request) {
	userAPIs, err := a.ResolveExternalAPIs()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
	}

	rtc := make([]externalAPI, 0, len(userAPIs))
	for typ, api := range userAPIs {
		conf := api.Config()

		rtc = append(rtc, externalAPI{
			Type:           typ,
			Name:           conf.Name,
			SupportedTypes: conf.WorkoutTypes,
		})
	}

	response.WriteJson(rtc, 200, w)
}

func (a *Api) ServiceTab(w http.ResponseWriter, r *http.Request) {
	_, apiType, ok := a.getApiByName(r.PathValue("type"))
	if !ok {
		errors.NotFound().Write(w, r)
		return
	}

	userAPIs, err := a.GetExternalAPIs()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	a.R().Tmpl.RenderDirect(a.externalApiPanel(apiType, userAPIs))
}

func (a *Api) Authenticate(w http.ResponseWriter, r *http.Request) {
	api, apiType, ok := a.getApiByName(r.PathValue("type"))
	if !ok {
		errors.NotFound().Write(w, r)
		return
	}

	rawStep := r.PathValue("step")
	stepIndex, err := strconv.Atoi(rawStep)
	if err != nil {
		errors.BadRequest("Invalid authentication step").Write(w, r)
		return
	}

	steps := api.Config().AuthenticationSteps
	if stepIndex < 0 || stepIndex >= len(steps) {
		errors.BadRequest("Invalid authentication step").Write(w, r)
		return
	}

	ui := externalapi.AuthenticationUI{
		R: a.R(),
		BuildForm: func(fields []form.Field, buttonTranslation string, requestedStepIndex int) templ.Component {
			return a.authForm(apiType, fields, requestedStepIndex, buttonTranslation)
		},
	}

	result, err := steps[stepIndex].Execute(r, ui)
	if err != nil {
		errors.BadRequest(err.Error()).Write(w, r)
		return
	}

	if result.Config != nil {
		if err := a.saveExternalAPI(result.Config); err != nil {
			err.GetErrorStruct().Write(w, r)
			return
		}

		userAPIs, dbErr := a.GetExternalAPIs()
		if dbErr != nil {
			dbErr.GetErrorStruct().Write(w, r)
			return
		}

		a.R().Tmpl.RenderDirect(a.serviceContent(apiType, userAPIs))
		return
	}

	if result.NextComponent != nil {
		a.R().Tmpl.RenderDirect(result.NextComponent)
		return
	}

	userAPIs, dbErr := a.GetExternalAPIs()
	if dbErr != nil {
		dbErr.GetErrorStruct().Write(w, r)
		return
	}

	a.R().Tmpl.RenderDirect(a.serviceContent(apiType, userAPIs))
}

func (a *Api) DeleteSession(w http.ResponseWriter, r *http.Request) {
	_, apiType, ok := a.getApiByName(r.PathValue("type"))
	if !ok {
		errors.NotFound().Write(w, r)
		return
	}

	if err := a.deleteExternalAPI(apiType); err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	userAPIs, err := a.GetExternalAPIs()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	a.R().Tmpl.RenderDirect(a.serviceContent(apiType, userAPIs))
}

func (a *Api) renderOverview(w http.ResponseWriter, r *http.Request, selected models.ExternalAPIType) {
	userAPIs, err := a.GetExternalAPIs()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	a.R().Tmpl.Render(a.overviewPage(selected, userAPIs), "generic.appName", "generic.appName")
}

func (a *Api) UploadWorkout(w http.ResponseWriter, r *http.Request) {
	api, _, ok := a.getApiByName(r.PathValue("type"))
	if !ok {
		errors.NotFound().Write(w, r)
		return
	}

	workoutId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("Invalid workout id").Write(w, r)
		return
	}

	if err := a.upload(api, workoutId); err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	response.WriteText("OK", 200, w)
}
