package statistics

import (
	"time"

	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// Select statement to get exact PAI values grouped by a moving average of 7 days
const selectPAIExact = `
SELECT
	yd.:idx AS id,
	ydStart.user_start_offset AS start,
	ydEnd.user_end_offset AS end,
	AVG(r.value) AS pai
FROM (
	SELECT 
		yd.:idx AS idx,
		yd.id AS day_idx,
		SUM(p.workout_pai + p.steps_pai) OVER (
			-- PARTITION BY p.user_id
			ORDER BY p.id
			ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
		) AS value,
		p.workout_pai + p.steps_pai AS earned
	FROM (
		SELECT 
			glob.id,
			NVL(SUM(w.pai), 0) AS workout_pai,
			NVL(SUM(s.pai), 0) As steps_pai
		FROM (
			SELECT off.*
			FROM year_day yda
			INNER JOIN v_year_day_user_offset off ON off.id = yda.id AND off.user_id = :user_id
			WHERE
  		  		yda.:idx >= (SELECT ydd.:idx FROM year_day ydd WHERE ydd.start >= :start ORDER BY ydd.start LIMIT 1) - :offset - 7
  	 		AND yda.:idx <= (SELECT ydd.:idx FROM year_day ydd WHERE ydd.start >= :start ORDER BY ydd.start LIMIT 1) + :offset
		) glob
		LEFT JOIN workout w ON w.user_id = :user_id AND w.start >= glob.user_start_offset AND w.start <= glob.user_end_offset
		LEFT JOIN steps_pai s ON s.id = glob.id AND s.user_id = :user_id
		GROUP BY glob.id
	) p
	INNER JOIN year_day yd ON yd.id = p.id
) r
INNER JOIN year_day yd ON yd.id = r.day_idx
INNER JOIN (
	SELECT MAX(yd.id) AS idUnit, yd.:idx
	FROM year_day yd
	GROUP BY yd.:idx
) ydMax ON ydMax.:idx = yd.:idx
INNER JOIN (
	SELECT MIN(yd.id) AS idUnit, yd.:idx
	FROM year_day yd
	GROUP BY yd.:idx
) ydMin ON ydMax.:idx = ydMin.:idx
INNER JOIN v_year_day_user_offset ydEnd ON ydEnd.id  = ydMax.idUnit AND ydEnd.user_id = :user_id
INNER JOIN v_year_day_user_offset ydStart ON ydStart.id = ydMin.idUnit AND ydStart.user_id = :user_id
-- The first seven days were only selected for calculation
WHERE yd.:idx >= (SELECT ydd.:idx FROM year_day ydd WHERE ydd.start >= :start ORDER BY ydd.start LIMIT 1) - 21
GROUP BY yd.:idx
ORDER BY yd.:idx ASC
`

// Select statement to get average PAI values grouped by the provided time range.
// This isn't totally exact but the execution is much faster on large datasets
const selectPAIAverage = `
SELECT 
	yd.:idx AS id,
	ydStart.start AS start,
	ydEnd.end AS end,
	SUM(NVL(w.pai, 0) + NVL(s.pai, 0)) / (COUNT(DISTINCT yd.id) / 7) AS pai
FROM (
	SELECT off.*
	FROM year_day yda
	INNER JOIN v_year_day_user_offset off ON off.id = yda.id AND off.user_id = :user_id
	WHERE
  		yda.:idx >= (SELECT ydd.:idx FROM year_day ydd WHERE ydd.start >= :start ORDER BY ydd.start LIMIT 1) - :offset
  	AND yda.:idx <= (SELECT ydd.:idx FROM year_day ydd WHERE ydd.start >= :start ORDER BY ydd.start LIMIT 1) + :offset
) yd
INNER JOIN (
	SELECT MAX(yd.id) AS idUnit, yd.:idx
	FROM year_day yd
	GROUP BY yd.:idx
) ydMax ON ydMax.:idx = yd.:idx
INNER JOIN (
	SELECT MIN(yd.id) AS idUnit, yd.:idx
	FROM year_day yd
	GROUP BY yd.:idx
) ydMin ON ydMax.:idx = ydMin.:idx
INNER JOIN v_year_day_user_offset ydEnd ON ydEnd.id  = ydMax.idUnit AND ydEnd.user_id = :user_id
INNER JOIN v_year_day_user_offset ydStart ON ydStart.id = ydMin.idUnit AND ydStart.user_id = :user_id
LEFT JOIN workout w ON w.user_id = :user_id AND w.start >= yd.user_start_offset AND w.start <= yd.user_end_offset
LEFT JOIN steps_pai s ON s.id = yd.id AND s.user_id = :user_id
GROUP BY yd.:idx
`

type paiData struct {
	statisticsRow

	// The number of steps taken in the time period
	PAI float64 `json:"pai"`
}

func (api *Api) getPAIData(center time.Time, unit SamplingUnit, cnt int) ([]paiData, errors.Error) {
	var sql string
	switch unit {
	case SamplingDay, SamplingWeek:
		sql = selectPAIExact
	case SamplingMonth, SamplingYear:
		sql = selectPAIAverage
	default:
		return nil, errors.InternalError().Log("Invalid sampling unit for PAI data request", nil, api)
	}

	sql = api.getCustomRangeSelect(sql, center, unit, cnt)

	rtc := []paiData{}
	if err := api.R().Db.QueryStructs(&rtc, sql); err != nil {
		return nil, err.GetResponse().Log("Failed to query workout data", err, api)
	}

	return api.transformPAIData(rtc, unit), nil
}

func (api *Api) transformPAIData(rows []paiData, unit SamplingUnit) []paiData {
	// Only add label to data
	for i, row := range rows {
		rows[i].Label, rows[i].LabelTooltip = unit.getLabel(row.Start, row.End)
		rows[i].Start = api.transformDate(rows[i].Start)
		rows[i].End = api.transformDate(rows[i].End)
	}

	return rows
}
