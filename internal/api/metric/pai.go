package metric

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// PaiDay describes the PAI's value for a specific
// weekday within a set of days (by default seven days)
type PaiDay struct {

	// Current PAI value
	Value int `db:"value" json:"value"`

	// Short abbrevation name of the weekday
	WeekdayShort string `json:"weekdayAbbrevation"`

	// Indexing of the weekday (0 = MONDAY, 1 = TUESDAY)
	WeekdayIndex int `db:"weekday_index" json:"weekdayIndex"`

	// How many PAIs were earned at this date
	Earned int `db:"earned" json:"earned"`

	// Unique and incrementing ID of this PAI day (days since unix epoch with client timezone offset applied)
	DayIndex int `db:"day_index" json:"dayIndex"`
}

// PaiProgression contains the current PAI score with
// the progression values of the last week
type PaiProgression struct {

	// Current PAI score
	Score int `json:"score"`

	// Progression over the last seven day
	Progression []PaiDay `json:"progression"`
}

// GetPaiProgression returns the current PAI score with the daily
// progression of the last week as details
func (a *Api) GetPaiProgression() (rtc PaiProgression, err errors.Error) {
	var wg sync.WaitGroup
	errChan := make(chan errors.Error)

	// Starting / ending date to calculate the PAI score from
	startDate := time.Now().AddDate(0, 0, -6)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endDate := time.Now().AddDate(0, 0, 1)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())

	// Number of functions
	wg.Add(2)

	// Get the current PAI score
	go func() {
		defer wg.Done()
		var dbError database.Error

		rtc.Score, dbError = a.GetSumOfPai(startDate, time.Now())
		if dbError != nil {
			errChan <- errors.InternalError().Log("Failed to query current PAI value: %s", dbError, a)
		}
	}()

	// Get daily PAI score
	go func() {
		defer wg.Done()

		if weekly, wErr := a.GetWeeklyPaiScore(startDate, endDate); wErr != nil {
			errChan <- wErr
		} else {
			rtc.Progression = weekly
		}
	}()

	// Read any error from channel
	go func() {
		for {
			errCh, ok := <-errChan

			// Channel closed
			if !ok {
				break
			} else {
				// Write to error variable
				err = errCh
			}
		}
	}()

	// Wait and return first error
	wg.Wait()
	close(errChan)

	return
}

// GetWeeklyPaiScore returns the calculated PAI score for the provided
// time range (should be weekly to match upstream PAI value)
func (a *Api) GetWeeklyPaiScore(startDate, endDate time.Time) (rtc []PaiDay, err errors.Error) {
	paiDaily, placeholders := a.getPaiSelect(startDate.AddDate(0, 0, -7), endDate, a.R().User.Id)

	dbError := a.R().Db.QueryStructs(&rtc, `
		SELECT r.weekday_index, r.day_index, r.value, r.earned
		FROM (
			SELECT 
				yd.day_week AS weekday_index, 
				ROUND((UNIX_TIMESTAMP(yd.start) + ?) / (24 * 60 * 60)) + 1 AS day_index,
				SUM(p.workout_pai + p.steps_pai) OVER (
					-- PARTITION BY p.user_id
					ORDER BY p.id
					ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
				) AS value,
				p.workout_pai + p.steps_pai AS earned,
				yd.start
			FROM (`+paiDaily+`) p
			LEFT JOIN year_day yd ON yd.id = p.id
		) r
		WHERE r.start >= ?
		ORDER BY r.start ASC
	`, append([]any{a.R().User.GetTimeZoneOffset()}, append(placeholders, startDate)...)...)

	if dbError != nil {
		return rtc, errors.InternalError().Log("Failed to query weekly PAI values: %s", dbError, a)
	}

	// Get name for weekday
	for i := range rtc {
		rtc[i].WeekdayShort = a.R().Tr.Get(fmt.Sprintf("weekday.short_%d", rtc[i].WeekdayIndex))
	}

	return
}

// getPaiSelect returns the select statement for the daily PAI values.
// Because we have to calculate all values based on the users time offset, no view (pai_daily)
// is used here for performance (~around 3x faster)
func (a *Api) getPaiSelect(startDate, endDate time.Time, userId int) (sql string, placeholder []any) {
	sql = `
		SELECT 
			glob.id,
			NVL(SUM(w.pai), 0) AS workout_pai,
			NVL(SUM(s.pai), 0) AS steps_pai
		FROM v_year_day_user_offset glob
		LEFT JOIN workout w ON w.user_id = :user_id AND w.start >= glob.user_start_offset AND w.start <= glob.user_end_offset
		LEFT JOIN steps_pai s ON s.id = glob.id AND s.user_id = :user_id
		WHERE 
			glob.start >= ?
		AND glob.end <= ?
		AND glob.user_id = :user_id
		GROUP BY glob.id
	`

	return strings.ReplaceAll(sql, ":user_id", strconv.Itoa(userId)), []any{
		startDate, endDate,
	}
}

// GetSumOfPai returns the PAI score within the provided time range
func (a *Api) GetSumOfPai(startDate, endDate time.Time) (rtc int, dbError database.Error) {
	paiDaily, placeholders := a.getPaiSelect(startDate, endDate, a.R().User.Id)

	dbError = a.R().Db.QueryForValue(&rtc, `
		SELECT NVL(SUM(pd.workout_pai + pd.steps_pai), 0)
		FROM (`+paiDaily+`) pd
	`, placeholders...)

	return
}

// cacheStepsPAI calculates the PAI points for the provided user and time range
// and stores them within the cache table
func (a *Api) cacheStepsPAI(start, end time.Time, userID int) error {
	// Select one more day to make sure we have all data for a single day
	start = start.AddDate(0, 0, -1)
	end = end.AddDate(0, 0, 1)

	// Use a transaction to delete any existing data withing the range
	trans, err := a.R().Db.NewTransaction()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	_, err = trans.Db.Exec(`
		DELETE p FROM steps_pai p
		INNER JOIN v_year_day_user_offset yd ON yd.user_id = ? AND yd.id = p.id
		WHERE p.user_id = ?
		  AND yd.start > ?
		  AND yd.end < ?`,
		userID, userID, start, end,
	)
	if err != nil {
		trans.RollbackTransaction()
		return fmt.Errorf("failed to delete existing PAI cache data: %w", err)
	}

	sel := `
		INSERT INTO steps_pai (id, user_id, pai)
		SELECT
			ydd.id,
			:user_id AS user_id,
			(CASE
				WHEN NVL(SUM(s.count), 0) - NVL(SUM(w.count), 0) >= 30000 THEN 10
				WHEN NVL(SUM(s.count), 0) - NVL(SUM(w.count), 0) >= 20000 THEN 5
				WHEN NVL(SUM(s.count), 0) - NVL(SUM(w.count), 0) >= 10000 THEN 2
				ELSE 0
			END) pai
		FROM year_day ydd
		-- This isn't totally correct because a workout could not have steps tracked. But we can't relay
		-- on the start and end time of the workout because no steps in the pauses / when workout got merged
		-- are counted. Checking the workout details would be too slow so this is the only solution
		LEFT JOIN (
			SELECT
				yd.id,
				SUM(s.count) AS count
			FROM v_year_day_user_offset yd
			INNER JOIN steps s ON s.user_id = :user_id AND s.start >= yd.user_start_offset AND s.start <= yd.user_end_offset
			-- Do not select steps that were made during workouts. Only use workouts with small pauses because we want to include steps that were made when the workout was paused / ended
			LEFT JOIN workout w ON s.start > w.start AND s.start < w.end AND w.user_id = s.user_id AND TIMESTAMPDIFF(SECOND, w.start, w.end) - w.duration < w.duration * 0.1 AND w.steps = 0
			WHERE yd.user_id = :user_id AND w.id IS NULL
			GROUP BY yd.id
		) s ON ydd.id = s.id
		LEFT JOIN (
			SELECT
				yd.id,
				SUM(w.steps) AS count
			FROM v_year_day_user_offset yd
			INNER JOIN workout w ON w.user_id = :user_id AND w.start >= yd.user_start_offset AND w.start <= yd.user_end_offset
			GROUP BY yd.id
		) w on ydd.id = w.id
		WHERE ydd.start > ? AND ydd.end < ?
		GROUP BY ydd.id 
	`
	sel = strings.ReplaceAll(sel, ":user_id", fmt.Sprintf("%d", userID))
	if _, err := trans.Db.Exec(sel, start, end); err != nil {
		trans.RollbackTransaction()
		return fmt.Errorf("failed to insert new PAI cache data: %w", err)
	}

	if err := trans.CommitTransaction(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
