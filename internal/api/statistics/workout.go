package statistics

import (
	"database/sql"
	"math"
	"strings"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/api/workout/shared"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

type workoutData struct {
	statisticsRow

	Distance  map[int]float64 `json:"distance"`
	Calories  map[int]float64 `json:"calories"`
	Duration  map[int]float64 `json:"duration"`
	PAI       map[int]float64 `json:"pai"`
	Count     map[int]float64 `json:"count"`
	Speed     map[int]float64 `json:"speed"`
	Heartrate map[int]float64 `json:"heartrate"`

	// The amount of rows we already received to calculate a moving average
	rowCnt int
}

type workoutRow struct {
	UnitID     int
	Start      time.Time
	End        time.Time
	Distance   float64
	Calories   float64
	Duration   float64
	PaiWorkout float64
	Count      float64
	Heartrate  float64
	Speed      float64
	TypeID     sql.NullInt32
}

func (api *Api) getWorkoutData(center time.Time, unit SamplingUnit, aggregation AggregateFunction, cnt int, filter *shared.WorkoutFilter) ([]workoutData, errors.Error) {
	rows := []workoutRow{}

	// Get filtered workouts
	sel := api.R().Db.Struct.QuerySlice(&rows)
	sel.Where().Column(models.Workout_UserId, "=", api.R().User.Id).Add()

	// Apply filter values
	if err := shared.ApplyFilter(filter, sel); err != nil {
		return nil, err
	}

	// We only extract the where placeholders (because of the internal logic we want to use)
	// from the struct.
	// We add it inside the join condition because otherweise the execution plan will build
	// a join buffer
	whereSQL, wherePlaceholder := sel.GetWhereStatement()
	whereSQL = strings.ReplaceAll(whereSQL, "workout.", "w.")
	whereSQL = strings.ReplaceAll(whereSQL, "w.w.", "w.")
	baseSelect := api.getRangeSelect(center, unit, cnt)
	sqll := `
		SELECT 
			units.idx AS unitId,
			units.start_utc AS start,
			units.end_utc AS end,
			NVL(:agg(w.distance), 0) AS distance,
			NVL(:agg(w.calories - calories_default), 0) AS calories,
			NVL(:agg(w.duration), 0) AS duration,
			NVL(:agg(w.pai), 0) AS pai_workout,
			NVL(AVG(w.speed_av), 0) AS speed,
			NVL(AVG(w.heart_rate_av), 0) AS heartrate,
			w.type_id,
			COUNT(*) AS count
		FROM (` + baseSelect + `) units
		LEFT JOIN workout w ON w.start >= units.start AND w.start <= units.end
		  ` + whereSQL + `
		GROUP BY units.idx, units.start, units.end, w.type_id
		ORDER BY units.idx
	`
	sqll = strings.ReplaceAll(sqll, ":agg", aggregation.GetForSQL())
	if err := api.R().Db.QueryStructs(&rows, sqll, wherePlaceholder...); err != nil {
		return []workoutData{}, err.GetResponse().Log("Failed to query workout data", err, api)
	}

	return api.transformWorkoutRows(rows, aggregation, unit, cnt), nil
}

func (api *Api) transformWorkoutRows(rows []workoutRow, aggregation AggregateFunction, unit SamplingUnit, cnt int) []workoutData {
	// +1 to make sure we have enough capacity for odd numbers
	rtc := make([]workoutData, 0, cnt+1)

	currentData := newWorkoutData()
	for _, row := range rows {
		// New data
		if currentData.ID != 0 && currentData.ID != row.UnitID {
			rtc = append(rtc, currentData)
			currentData = newWorkoutData()
		}

		// Fill statistic data which should be the same for all workouts
		if currentData.ID == 0 {
			currentData.statisticsRow = statisticsRow{
				Start: api.transformDate(row.Start),
				End:   api.transformDate(row.End),
				ID:    row.UnitID,
			}
			currentData.Label, currentData.LabelTooltip = unit.getLabel(row.Start, row.End)
		}

		typ := getWorkoutTypeIndex(row.TypeID)
		currentData.Calories[typ] = row.Calories
		currentData.Count[typ] = row.Count
		currentData.Distance[typ] = row.Distance
		currentData.Duration[typ] = row.Duration
		currentData.PAI[typ] = row.PaiWorkout
		currentData.Heartrate[typ] = row.Heartrate
		currentData.Speed[typ] = row.Speed

		buildTotal(typ, &row, &currentData, aggregation)
	}

	if currentData.ID != 0 {
		rtc = append(rtc, currentData)
	}

	for i, d := range rtc {
		for key, val := range d.Speed {
			if val > 7200 || val <= 0.01 {
				rtc[i].Speed[key] = 0
			} else {
				rtc[i].Speed[key] = math.Round((3600/val)*1000) / 1000.0
			}
		}
	}

	return rtc
}

func newWorkoutData() workoutData {
	return workoutData{
		Calories:  make(map[int]float64),
		Count:     make(map[int]float64),
		Distance:  make(map[int]float64),
		Duration:  make(map[int]float64),
		PAI:       make(map[int]float64),
		Speed:     make(map[int]float64),
		Heartrate: make(map[int]float64),
	}
}

func getWorkoutTypeIndex(idx sql.NullInt32) int {
	// No specific workout data available => value is also probably null
	if !idx.Valid {
		return -1
	}

	return int(idx.Int32)
}

// buildTotal calculates the total / average values based on the provided
// row data
func buildTotal(typeID int, row *workoutRow, data *workoutData, aggregation AggregateFunction) {
	if typeID == -1 {
		return
	}

	data.rowCnt += 1

	if aggregation == AggregateFunctionSum {
		data.Calories[-1] += row.Calories
		data.Distance[-1] += row.Distance
		data.Duration[-1] += row.Duration
		data.PAI[-1] += row.PaiWorkout
	} else {
		data.Calories[-1] += (row.Calories - data.Calories[-1]) / float64(data.rowCnt)
		data.Distance[-1] += (row.Distance - data.Distance[-1]) / float64(data.rowCnt)
		data.Duration[-1] += (row.Duration - data.Duration[-1]) / float64(data.rowCnt)
		data.PAI[-1] += (row.PaiWorkout - data.PAI[-1]) / float64(data.rowCnt)
	}

	// It doesn't make sense to build average of cound
	data.Count[-1] += row.Count

	// It doesn't make sense to sum up these values
	data.Heartrate[-1] += (row.Heartrate - data.Heartrate[-1]) / float64(data.rowCnt)
	data.Speed[-1] += (row.Speed - data.Speed[-1]) / float64(data.rowCnt)
}
