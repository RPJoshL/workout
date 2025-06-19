package dashboard

import (
	"math"
	"sync"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/api/metric"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// DashboardData contains data needed for displaying
// the dashboard site
type DashboardData struct {

	// Current (past 7 days) PAI score
	CurrentPaiScore int

	// Weeklay PAI values beginning seven days ago
	WeeklyPaiScore []metric.PaiDay

	DailySteps     int
	WeeklyStepsAvg int
}

// GetDashboardData fetches all data needed for the dashboard page
func (a *Api) GetDashboardData() (rtc DashboardData, err errors.Error) {
	progression, err := a.Metric.GetPaiProgression()
	rtc.CurrentPaiScore = progression.Score
	rtc.WeeklyPaiScore = progression.Progression

	var wg sync.WaitGroup
	errChan := make(chan errors.Error)

	// Starting / ending date to calculate the PAI score from
	startDate := time.Now().AddDate(0, 0, -6)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endDate := time.Now().AddDate(0, 0, 1)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())

	// Number of functions
	wg.Add(4)

	// Get the current PAI score
	go func() {
		defer wg.Done()
		var dbError database.Error

		rtc.CurrentPaiScore, dbError = a.Metric.GetSumOfPai(startDate, endDate)
		if dbError != nil {
			errChan <- errors.InternalError().Log("Failed to query current PAI value: %s", dbError, a)
		}
	}()

	// Get daily PAI score
	go func() {
		defer wg.Done()

		if weekly, wErr := a.Metric.GetWeeklyPaiScore(startDate, endDate); wErr != nil {
			errChan <- wErr
		} else {
			rtc.WeeklyPaiScore = weekly
		}
	}()

	// Get step values
	go func() {
		defer wg.Done()
		var err errors.Error

		// Get start of day in users time zone
		startDate := time.Now().In(a.R().User.TimeZone)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, a.R().User.TimeZone)

		rtc.DailySteps, err = a.Metric.GetStepsSince(startDate.UTC())
		if err != nil {
			errChan <- err
		}
	}()
	go func() {
		defer wg.Done()
		var err errors.Error

		rtc.WeeklyStepsAvg, err = a.Metric.GetStepsSince(time.Now().Add(-7 * 24 * time.Hour))
		if err != nil {
			errChan <- err
		} else {
			rtc.WeeklyStepsAvg = int(math.Round(float64(rtc.WeeklyStepsAvg) / 7.0))
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
