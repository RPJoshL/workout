package statistics

import (
	"net/http"
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
	AggregationStr  string `query:"aggregation"`
	SamplingUnitStr string `query:"samplingUnitStr"`
}

func GetRoutes() *router.Router {
	api := &Api{}

	routes := router.Routes{
		router.NewRoute(
			"StatisticView",
			"GET",
			"/",
			api.GetStatistics,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) GetStatistics(w http.ResponseWriter, r *http.Request) {
	filter := &statisticRequestApi{}
	if err := api.R().Parser.Parse(filter, router.RequestParserOptions{}); err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	switch filter.AggregationStr {
	case "", "SUM":
		filter.Aggregation = AggregateFunctionSum
	case "SVG":
		filter.Aggregation = AggregateFunctionAvg
	default:
		errors.BadRequest("Unknown aggregation").Write(w, r)
		return
	}

	switch filter.SamplingUnitStr {
	case "", "DAY":
		filter.SamplingUnit = SamplingDay
	case "WEEK":
		filter.SamplingUnit = SamplingWeek
	case "MONTH":
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
	if filter.CenterTime.IsZero() {
		filter.CenterTime = time.Now().Add((-24 * time.Hour) * (time.Duration(filter.Count / 2)))
	}

	data, err := api.getStatisticData(&filter.statisticRequest)
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	api.R().Tmpl.Render(api.main(data), "generic.appName", "generic.appName")
}
