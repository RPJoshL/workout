package statistics

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

type SamplingUnit int8

const (
	SamplingDay SamplingUnit = iota
	SamplingWeek
	SamplingMonth
	SamplingYear
)

type AggregateFunction int8

const (
	AggregateFunctionSum AggregateFunction = iota
	AggregateFunctionAvg
)

var samplingUnitToYearDayColumn = map[SamplingUnit]string{
	SamplingDay:   "id",
	SamplingWeek:  "week_id",
	SamplingMonth: "month_id",
	SamplingYear:  "year_id",
}

type statisticsRow struct {
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
	ID           int       `json:"id"`
	Label        string    `json:"label"`
	LabelTooltip string    `json:"labelTooltip"`
}

type statisticRequest struct {
	shared.WorkoutFilter

	CenterTime   time.Time
	Count        int `query:"count"`
	Aggregation  AggregateFunction
	SamplingUnit SamplingUnit
}

func (api *Api) getStatisticData(req *statisticRequest) (StatisticPageData, errors.Error) {
	rtc := StatisticPageData{
		CenterDate: req.CenterTime,
	}

	var wg sync.WaitGroup
	var rtcError errors.Error
	wg.Add(5)
	go func() {
		if err := api.R().Db.Struct.QuerySlice(&rtc.Types).Run(); err != nil {
			api.Logger().Error("Failed to query workout types: %s", err)
		}
		wg.Done()
	}()
	go func() {
		if err := api.R().Db.Struct.QuerySlice(&rtc.Tags).Run(); err != nil {
			api.Logger().Error("Failed to query workout tags: %s", err)
		}
		wg.Done()
	}()
	go func() {
		rtc.WorkoutData = []workoutData{}
		data, err := api.getWorkoutData(req.CenterTime, req.SamplingUnit, req.Aggregation, req.Count, &req.WorkoutFilter)
		if err != nil {
			rtcError = err
			api.Logger().Error("Failed to query workout data: %s", err)
		} else {
			rtc.WorkoutData = data
		}
		wg.Done()
	}()
	go func() {
		rtc.StepData = []stepData{}
		data, err := api.getStepData(req.CenterTime, req.SamplingUnit, req.Aggregation, req.Count)
		if err != nil {
			rtcError = err
			api.Logger().Error("Failed to query step data: %s", err)
		} else {
			rtc.StepData = data
		}
		wg.Done()
	}()
	go func() {
		rtc.PaiData = []paiData{}
		data, err := api.getPAIData(req.CenterTime, req.SamplingUnit, req.Count)
		if err != nil {
			rtcError = err
			api.Logger().Error("Failed to query PAI data: %s", err)
		} else {
			rtc.PaiData = data
		}
		wg.Done()
	}()

	wg.Wait()

	return rtc, rtcError
}

func (a AggregateFunction) GetForSQL() string {
	switch a {
	case AggregateFunctionAvg:
		return "AVG"
	case AggregateFunctionSum:
		return "SUM"
	default:
		return "SUM"
	}
}

func (s SamplingUnit) getLabel(start, end time.Time) (def, label string) {
	switch s {
	case SamplingDay:
		return start.Format("02.01"), start.Format("02.01.06")
	case SamplingWeek:
		_, week := start.ISOWeek()
		return fmt.Sprintf("%02d", week), fmt.Sprintf("%s - %s (%02d)", start.Format("02.01"), end.Format("02.01"), week)
	case SamplingMonth:
		return start.Format("01.06"), start.Format("01.2006")
	case SamplingYear:
		return start.Format("06"), start.Format("2006")
	default:
		return "", ""
	}
}

// getRangeSelect returns a select statement to select all ranges (eg. 01.01.2001 - 07.01.2001)
// centered by the provided day with the given cnt. If cnt is not dividable by 2, an additional
// row is returned
func (api *Api) getRangeSelect(centerDate time.Time, unit SamplingUnit, cnt int) string {
	rtc := `
		SELECT
			ydStart.:idx AS idx,
			ydStart.user_start_offset AS start,
			ydEnd.user_end_offset AS end,
			ydStart.start AS start_utc,
			ydEnd.end AS end_utc
		FROM (
			SELECT MAX(yd.id) AS idUnit, yd.:idx
			FROM year_day yd
			GROUP BY yd.:idx
		) ydMax
		INNER JOIN (
			SELECT MIN(yd.id) AS idUnit, yd.:idx
			FROM year_day yd
			GROUP BY yd.:idx
		) ydMin ON ydMax.:idx = ydMin.:idx
		INNER JOIN v_year_day_user_offset ydEnd ON ydEnd.id  = ydMax.idUnit AND ydEnd.user_id = :user_id
		INNER JOIN v_year_day_user_offset ydStart ON ydStart.id = ydMin.idUnit AND ydStart.user_id = :user_id
		WHERE
			ydStart.:idx >= (SELECT ydd.:idx FROM year_day ydd WHERE ydd.start >= :start ORDER BY ydd.start LIMIT 1) - :offset
		AND  ydStart.:idx <= (SELECT ydd.:idx FROM year_day ydd WHERE ydd.start >= :start ORDER BY ydd.start LIMIT 1) + :offset
	`

	return api.getCustomRangeSelect(rtc, centerDate, unit, cnt)
}

func (api *Api) getCustomRangeSelect(sql string, centerDate time.Time, unit SamplingUnit, cnt int) string {
	samplingColumn, ok := samplingUnitToYearDayColumn[unit]
	if !ok {
		return ""
	}

	sql = strings.ReplaceAll(sql, ":idx", samplingColumn)
	sql = strings.ReplaceAll(sql, ":start", "'"+centerDate.Format("2006-01-02")+"T00:00:00'")
	sql = strings.ReplaceAll(sql, ":user_id", strconv.Itoa(api.R().User.Id))

	// It's okay when we do have odd number (3 = 2 => xMx)
	offset := cnt / 2
	sql = strings.ReplaceAll(sql, ":offset", strconv.Itoa(offset))

	return sql
}

func getDefaultCenterDate(unit SamplingUnit, cnt int) time.Time {
	switch unit {
	case SamplingDay:
		return time.Now().Add((-24 * time.Hour) * (time.Duration(cnt / 2)))
	case SamplingWeek:
		return time.Now().Add((-24 * time.Hour) * (time.Duration(cnt / 2)) * 7)
	case SamplingMonth:
		now := time.Now()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, (cnt/2)*-1, 0)
	case SamplingYear:
		now := time.Now()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate((cnt/2)*-1, 0, 0)
	default:
		return time.Now()
	}
}

// transformDate transforms an existing time received by the statistic
// query to the users locale time
func (api *Api) transformDate(d time.Time) time.Time {
	return time.Date(
		d.Year(), d.Month(), d.Day(),
		d.Hour(), d.Minute(), d.Second(), d.Nanosecond(),
		api.R().User.TimeZone,
	)
}
