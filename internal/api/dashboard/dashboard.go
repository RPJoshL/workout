package dashboard

import (
	"fmt"
	"sync"
	"time"

	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// DashboardData contains data needed for displaying
// the dashboard site
type DashboardData struct {

	// Current (past 7 days) PAI score
	CurrentPaiScore int

	// Weeklay PAI values beginning seven days ago
	WeeklyPaiScore []PaiDay
}

// PaiDay describes the PAI's value for a specific
// weekday withn the last seven days
type PaiDay struct {

	// Current PAI value
	Value int `db:"value"`

	// Short abbrevation name of the weekday
	WeekdayShort string

	// Indexing of the weekday (0 = MONDAY, 1 = TUESDAY)
	WeekdayIndex int `db:"weekday_index"`

	// How many PAIs were earned at this date
	Earned int `db:"earned"`
}

// GetDashboardData fetches all data needed for the dashbaord page
func (a *Api) GetDashboardData() (rtc DashboardData, err errors.Error) {
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

		dbError := a.R().Db.QueryForValue(&rtc.CurrentPaiScore, `
			SELECT SUM(w.pai) FROM workout w
			WHERE w.start > ? AND w.user_id = ?
		`, startDate, a.R().User.Id)

		if dbError != nil {
			errChan <- errors.InternalError().Log("Failed to query current PAI value: %s", dbError, a)
		}
	}()

	// Get daily PAI score
	go func() {
		defer wg.Done()

		if weekly, wErr := a.getWeeklyPaiScore(startDate, endDate); wErr != nil {
			errChan <- wErr
		} else {
			rtc.WeeklyPaiScore = weekly
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

// getWeeklyPaiScore returns the calculated PAI score for the provided
// time range
func (a *Api) getWeeklyPaiScore(startDate, endDate time.Time) (rtc []PaiDay, err errors.Error) {
	dbError := a.R().Db.QueryStructs(&rtc, `
	SELECT
		WEEKDAY(CURRENT_DATE - INTERVAL i DAY) AS weekday_index,
		NVL ((  
			SELECT SUM(w.pai) from workout w
			WHERE w.start > ? - INTERVAL i DAY
			AND   w.start < ? - INTERVAL i DAY
			AND   w.user_id = ?
		), 0) AS value,
		NVL((  
			SELECT SUM(w.pai) from workout w
			WHERE w.start > ? - INTERVAL (1 + i) DAY
			AND   w.start < ? - INTERVAL i DAY
			AND   w.user_id = ?
		), 0) AS earned
	FROM 
	(
		SELECT 0 AS i UNION SELECT 1 AS i UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6
	) AS offsets
	ORDER BY offsets.i DESC
`, startDate, endDate, a.R().User.Id, endDate, endDate, a.R().User.Id)

	if dbError != nil {
		return rtc, errors.InternalError().Log("Failed to query weekly PAI values: %s", dbError, a)
	}

	// Get name for weekday
	for i := range rtc {
		rtc[i].WeekdayShort = a.R().Tr.Get(fmt.Sprintf("weekday.short_%d", rtc[i].WeekdayIndex))
	}

	return
}
