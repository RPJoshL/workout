package statistics

import (
	"time"

	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

type stepData struct {
	statisticsRow

	// The number of steps taken in the time period
	Steps int `json:"steps"`
}

func (a *Api) getStepData(center time.Time, unit SamplingUnit, aggregation AggregateFunction, cnt int) ([]stepData, errors.Error) {
	aggSql := `NVL(SUM(s.count), 0)`
	if aggregation == AggregateFunctionAvg {
		// Select average per day
		aggSql = `ROUND(NVL(SUM(s.count), 0) / DATEDIFF(units.end, units.start), 0)`
	}

	baseSelect := a.getRangeSelect(center, unit, cnt)
	sql := `
	SELECT 
		units.idx AS id,
		units.start_utc AS start,
		units.end_utc AS end,
		` + aggSql + ` AS steps
	FROM ( ` + baseSelect + ` ) units
	LEFT JOIN steps s ON s.start >= units.start AND s.start <= units.end AND s.user_id = ?
	GROUP By units.idx, units.start, units.end
	ORDER BY units.idx
	`

	rtc := []stepData{}
	if err := a.R().Db.QueryStructs(&rtc, sql, a.R().User.Id); err != nil {
		return nil, err.GetResponse().Log("Failed to query workout data", err, a)
	}

	return a.transformStepData(rtc, unit), nil
}

func (a *Api) transformStepData(rows []stepData, unit SamplingUnit) []stepData {
	// Only add label to data
	for i, row := range rows {
		rows[i].Label, rows[i].LabelTooltip = unit.getLabel(row.Start, row.End)
		rows[i].Start = a.transformDate(rows[i].Start)
		rows[i].End = a.transformDate(rows[i].End)
	}

	return rows
}
