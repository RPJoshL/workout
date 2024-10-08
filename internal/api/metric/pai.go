package metric

import (
	"fmt"
	"sync"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/database"
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
		var dbError database.DatabaseError

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
	dbError := a.R().Db.QueryStructs(&rtc, `
	SELECT
		WEEKDAY(CURRENT_DATE - INTERVAL i DAY) AS weekday_index,
		ROUND( (? + time.off) / (24 * 60 * 60) ) - i AS day_index,
		NVL ((  
			SELECT SUM(w.pai) from workout w
			WHERE w.start > ? - INTERVAL i DAY + INTERVAL time.off SECOND
			AND   w.start < ? - INTERVAL i DAY + INTERVAL time.off SECOND
			AND   w.user_id = ?
		), 0) AS value,
		NVL((  
			SELECT SUM(w.pai) from workout w
			WHERE w.start > ? - INTERVAL (1 + i) DAY + INTERVAL time.off SECOND
			AND   w.start < ? - INTERVAL i DAY + INTERVAL time.off SECOND
			AND   w.user_id = ?
		), 0) AS earned
	FROM 
	(
		SELECT 0 AS i UNION SELECT 1 AS i UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6
	) AS offsets, ( SELECT ? AS off ) AS time
	ORDER BY offsets.i DESC
`, endDate.Unix(), startDate, endDate, a.R().User.Id, endDate, endDate, a.R().User.Id, a.R().User.GetTimeZoneOffset())

	if dbError != nil {
		return rtc, errors.InternalError().Log("Failed to query weekly PAI values: %s", dbError, a)
	}

	// Get name for weekday
	for i := range rtc {
		rtc[i].WeekdayShort = a.R().Tr.Get(fmt.Sprintf("weekday.short_%d", rtc[i].WeekdayIndex))
	}

	return
}

// GetSumOfPai rturns the PAI score within the provided time range
func (a *Api) GetSumOfPai(startDate, endDate time.Time) (rtc int, dbError database.DatabaseError) {
	dbError = a.R().Db.QueryForValue(&rtc, `
		SELECT NVL(SUM(w.pai), 0) FROM workout w
		WHERE w.start > ? AND w.end < ?
		  AND w.user_id = ?
	`, startDate, endDate, a.R().User.Id)

	return
}
