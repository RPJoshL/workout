package statistics

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

type stepData struct {
	statisticsRow

	// The number of steps taken in the time period
	Steps int `json:"steps"`
}

func (api *Api) getStepData(req *statisticRequest) ([]stepData, errors.Error) {
	start, end := api.getSumDates(&req.WorkoutFilter, req.SamplingUnit)
	baseSelect := api.getRangeSelect(req.CenterTime, req.SamplingUnit, req.Count)

	sql := `
	SELECT 
		units.idx AS id,
		units.start_utc AS start,
		units.end_utc AS end,
		` + api.getAggregationStepSelect(req, start, end) + ` AS steps
	FROM ( ` + baseSelect + ` ) units
	LEFT JOIN steps s ON 
		 s.start >= units.start AND s.start <= units.end AND s.user_id = ?
	 AND s.start >= ? AND s.start <= ?
	GROUP By units.idx, units.start, units.end
	ORDER BY units.idx`

	rtc := []stepData{}
	if err := api.R().Db.QueryStructs(&rtc, sql, api.R().User.Id, start, end); err != nil {
		return nil, err.GetResponse().Log("Failed to query workout data", err, api)
	}

	return api.transformStepData(rtc, req.SamplingUnit), nil
}

func (api *Api) getAggregationStepSelect(req *statisticRequest, start, end time.Time) string {
	if req.Aggregation == AggregateFunctionSum {
		return `NVL(SUM(s.count), 0)`
	}

	// If the user selected a range, we can calculate the average over that range. We ignore days without steps tracked
	// as we want to show the user the "real" average
	if start.Year() >= 2000 && req.Aggregation == AggregateFunctionAvg && req.SamplingUnit == SamplingTotal {
		diffHours := end.Sub(start).Hours()
		return fmt.Sprintf("ROUND(NVL(SUM(s.count), 0) / %f)", diffHours/24.0)
	}

	// Get count of days which do have at least a single step tracked.
	// We don't want to calculate average with days where the tracker wasn't
	// available or worn by the user
	return `NVL(ROUND(NVL(SUM(s.count), 0) / NVL((
		SELECT
			COUNT(DISTINCT DAYOFYEAR(cntSteps.start) * YEAR(cntSteps.start))
		FROM steps cntSteps
		WHERE cntSteps.user_id = ` + strconv.Itoa(api.R().User.Id) + `
		  AND cntSteps.start >= units.start
		  AND cntSteps.start <= units.end
	), 1)), 0)`
}

// getSumDates returns the start and end constraints based on the provided duration of the filter
func (api *Api) getSumDates(filter *shared.WorkoutFilter, samplingUnit SamplingUnit) (start, end time.Time) {
	start = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	end = time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)

	if filter.DateRange == "" || samplingUnit != SamplingTotal {
		return
	}

	toIndex := strings.Index(filter.DateRange, " to ")
	t1, err1 := time.Parse("02.01.2006", filter.DateRange[0:toIndex])
	t2, err2 := time.Parse("02.01.2006", filter.DateRange[strings.LastIndex(filter.DateRange, " ")+1:])
	if err1 != nil || err2 != nil {
		api.Logger().Info("Date range for filtering is invalid: %s. Ignored it", filter.DateRange)
		return
	}

	return t1, t2
}

func (api *Api) transformStepData(rows []stepData, unit SamplingUnit) []stepData {
	// Only add label to data
	for i, row := range rows {
		rows[i].Label, rows[i].LabelTooltip = unit.getLabel(row.Start, row.End)
		rows[i].Start = api.transformDate(rows[i].Start)
		rows[i].End = api.transformDate(rows[i].End)
	}

	return rows
}
