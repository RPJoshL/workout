package statistics

import (
	"net/http"
	"strings"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/workout/cities"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

type Api struct {
	router.ApiRequest

	City cities.Api
}

type statisticRequestApi struct {
	statisticRequest

	CenterTimeStr   string `query:"centerTime"`
	AggregationStr  string `query:"aggregation"`
	SamplingUnitStr string `query:"samplingUnit"`
	DisplayCount    int    `query:"displayCount"`
}

func GetRoutes() *router.Router {
	api := &Api{}

	routes := router.Routes{
		router.NewRoute(
			"StatisticView",
			"GET",
			"/",
			api.GetStatisticPage,
			router.Options{},
		),
		router.NewRoute(
			"StatisticGraphData",
			"GET",
			"/graphData",
			api.GetStatisticGraphData,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) GetStatisticPage(w http.ResponseWriter, r *http.Request) {
	filter := &statisticRequestApi{
		statisticRequest: statisticRequest{
			Count:      60,
			CenterTime: getDefaultCenterDate(SamplingDay, 60),
		},
	}

	data, err := api.getStatisticData(&filter.statisticRequest)
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	api.R().Tmpl.Render(api.main(data), "generic.appName", "generic.appName")
}

func (api *Api) GetStatisticGraphData(w http.ResponseWriter, r *http.Request) {
	filter := &statisticRequestApi{}
	if err := api.R().Parser.Parse(filter, router.RequestParserOptions{}); err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	if filter.CenterTimeStr != "" {
		if tim, errA := time.Parse("02.01.2006", filter.CenterTimeStr); errA == nil {
			filter.CenterTime = tim
		} else {
			errors.BadRequest("#workout.dateInvalidFormat").Write(w, r)
			return
		}
	}

	switch strings.ToUpper(filter.AggregationStr) {
	case "", "SUM":
		filter.Aggregation = AggregateFunctionSum
	case "AVG":
		filter.Aggregation = AggregateFunctionAvg
	default:
		errors.BadRequest("Unknown aggregation").Write(w, r)
		return
	}

	switch strings.ToUpper(filter.SamplingUnitStr) {
	case "", "DAY":
		filter.SamplingUnit = SamplingDay
	case "WEEK":
		filter.SamplingUnit = SamplingWeek
	case "MONTH", "MON":
		filter.SamplingUnit = SamplingMonth
	case "YEAR":
		filter.SamplingUnit = SamplingYear
	default:
		errors.BadRequest("Unknown sampling unit").Write(w, r)
		return
	}

	// Use default values
	if filter.Count < 2 {
		filter.Count = 60
	}

	// Add soft maximum for years
	if filter.SamplingUnit == SamplingYear && filter.Count > 9 {
		filter.Count = 9
	}
	if filter.SamplingUnit == SamplingYear && filter.DisplayCount != 0 && filter.DisplayCount > 9 {
		filter.DisplayCount = 9
	}

	if filter.CenterTime.IsZero() {
		filter.CenterTime = getDefaultCenterDate(filter.SamplingUnit, filter.Count)
	}

	// When we cannot display all data for the provided display count (because it's too close to now),
	// we adjust the center date accordingly
	preferedCount := filter.Count
	if filter.DisplayCount > 0 {
		preferedCount = filter.DisplayCount
	}
	minDate := getDefaultCenterDate(filter.SamplingUnit, preferedCount)
	if filter.CenterTime.After(minDate) {
		filter.CenterTime = minDate
	}

	data, err := api.getStatisticData(&filter.statisticRequest)
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	api.R().Tmpl.RenderDirect(api.graphSection(data, false))
}
