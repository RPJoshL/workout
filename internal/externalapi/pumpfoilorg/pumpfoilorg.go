// Package pumpfoilorg handles the communication with the Pumpfoil.org
package pumpfoilorg

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/externalapi"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
	"github.com/RPJoshL/go-logger"
	"github.com/guregu/null/v5"
	"golang.org/x/sync/errgroup"
)

const (
	defaultURL         = "https://pumpfoil.org"
	pumpfoilAccelScale = 2048
	defaultAccelHz     = 25
	gpsChunkSize       = 500
	// How many chunks to upload in parallel
	parallelChunkCount = 10
	// How many accelerometer samples to use to determine the sample rate.
	// We expect that the rate is constant
	accelHzDetectSamples = 25
)

var _ externalapi.API = (*Api)(nil)

type Api struct{}

func (a *Api) Config() *externalapi.Configuration {
	return &externalapi.Configuration{
		DefaultURL: defaultURL,
		Name:       "Pumpfoil.org",
		IconURL:    "/static/img/logos/pumpfoilorg_icon.png",
		Type:       models.ExternalAPIPumpfoilOrg,
		AuthenticationSteps: []externalapi.AuthenticationStep{
			&authInit{},
			&authPoll{},
		},
		WorkoutTypes: []int{
			models.TYPE_PUMP_FOILING,
		},
	}
}

func (a *Api) UploadWorkout(workout *models.Workout, apiConfig *models.ExternalApi, marker externalapi.UploadMarker) error {
	startTime := time.Now()

	cred, err := parsePumpfoilOrgCredentials(apiConfig.Credentials)
	if err != nil {
		return err
	}

	client := NewPumpfoilOrgClient(apiConfig.ServerAddress, cred.DeviceToken)
	sessionUUID, err := utils.GenerateRandomString(32)
	if err != nil {
		return fmt.Errorf("unable to generate session ID: %w", err)
	}

	accelHz := detectAccelHz(workout)
	startReq := sessionStartRequest{
		SessionUUID: sessionUUID,
		StartedAt:   workout.Start.UTC().Format("2006-01-02T15:04:05Z"),
		Sport:       "pumpfoil",
		GpsHz:       1, // Lowest number we can use. Floating points are not supported
		AccelHz:     accelHz,
		AccelScale:  pumpfoilAccelScale,
		FoilID:      cred.FoilID,
	}

	if _, err := client.StartSession(startReq); err != nil {
		return fmt.Errorf("starting session: %w", err)
	}

	chunkIndex := 0

	gpsPoints := buildInterpolatedGpsPoints(workout)

	wg, ctx := errgroup.WithContext(context.Background())
	wg.SetLimit(parallelChunkCount)

	for i := 0; i < len(gpsPoints); i += gpsChunkSize {
		if ctx.Err() != nil {
			break
		}

		end := min(i+gpsChunkSize, len(gpsPoints))

		chunk := gpsPoints[i:end]
		chunkStartTime := gpsPoints[i][0].(int)

		start := i
		currentIndex := chunkIndex
		wg.Go(func() error {
			_, err := client.UploadChunk(sessionUUID, chunkRequest{
				Index:    currentIndex,
				Kind:     "gps",
				Encoding: "json",
				T0Ms:     chunkStartTime,
				Count:    len(chunk),
				Data:     chunk,
			})
			if err != nil {
				return fmt.Errorf("upload GPS chunk from index %d to %d: %w", start, end, err)
			}

			return nil
		})

		chunkIndex++
	}
	gpsChunks := chunkIndex

	if err := wg.Wait(); err != nil {
		return err
	}

	wg = &errgroup.Group{}
	wg.SetLimit(parallelChunkCount)

	// One accel chunk per workout-details bucket, so each chunk has a correct t0_ms.
	for _, point := range workout.WorkoutDetails {
		samples := point.GetAcceleration()
		if len(samples) < 4 {
			continue
		}

		xyz := make([]int16, 0, len(samples)/4*3)
		for i := 0; i < len(samples); i += 4 {
			xyz = append(xyz, samples[i+1], samples[i+2], samples[i+3])
		}

		// Es bringt nichts samples[0] rauszuwerfen
		t0Ms := int(samples[0]) + int(point.Time.Sub(workout.Start).Milliseconds())

		a.uploadAccelChunk(client, sessionUUID, wg, xyz, t0Ms, chunkIndex)
		chunkIndex++
	}

	if err := wg.Wait(); err != nil {
		return err
	}

	logger.Info("Uploaded %d (GPS) and %d (Acell) chunks to pumpfoil.org (accel_hz=%d) in %dms", gpsChunks, chunkIndex-gpsChunks, accelHz, time.Since(startTime).Milliseconds())

	res, err := client.CompleteSession(sessionUUID, SessionCompleteRequest{
		EndedAt:     workout.End.UTC().Format("2006-01-02T15:04:05Z"),
		TotalChunks: chunkIndex,
	})
	if err != nil {
		return fmt.Errorf("completing session: %w", err)
	}

	workout.PumpfoilorgSync = null.StringFrom(client.getSessionPath(res.SessionID))
	marker(workout, models.Workout_PumpfoilorgSync)

	return nil
}

func (a *Api) uploadAccelChunk(
	client *client, sessionUUID string, wg *errgroup.Group,
	data []int16, t0Ms int, chunkIdx int,
) {
	wg.Go(func() error {
		buf := make([]byte, len(data)*2)
		for k, v := range data {
			binary.LittleEndian.PutUint16(buf[k*2:], uint16(v))
		}

		_, err := client.UploadChunk(sessionUUID, chunkRequest{
			Index:    chunkIdx,
			Kind:     "accel",
			Encoding: "int16-b64",
			T0Ms:     t0Ms,
			Count:    len(data) / 3,
			Data:     base64.StdEncoding.EncodeToString(buf),
		})

		if err != nil {
			return fmt.Errorf("uploding acceleration chunk starting with idx %d: %w", chunkIdx, err)
		}

		return nil
	})
}

// detectAccelHz estimates the accelerometer sample rate from the first
// accelHzDetectSamples of the first workout point that contains acceleration data.
func detectAccelHz(workout *models.Workout) int {
	minLen := accelHzDetectSamples * 4

	for _, point := range workout.WorkoutDetails {
		samples := point.GetAcceleration()
		if len(samples) < minLen {
			continue
		}

		t0 := int(samples[0])
		tLast := int(samples[(accelHzDetectSamples-1)*4])

		dt := tLast - t0
		if dt <= 0 {
			continue
		}

		avgMs := float64(dt) / float64(accelHzDetectSamples-1)
		hz := int(math.Round(1000 / avgMs))
		if hz < 1 {
			return defaultAccelHz
		}

		return hz
	}

	return defaultAccelHz
}

// buildInterpolatedGpsPoints upsamples sparse GPS track points to one sample per second
// by linearly interpolating latitude, longitude and speed between consecutive points.
//
// This is required on the server side, so pumps are tracked correctly. But we don't
// want to track points more accurate than 3 seconds
func buildInterpolatedGpsPoints(workout *models.Workout) [][]any {
	validPoints := make([]models.WorkoutDetails, 0, len(workout.WorkoutDetails))
	for _, point := range workout.WorkoutDetails {
		if point.Latitude == 0 && point.Longitude == 0 {
			continue
		}

		validPoints = append(validPoints, point)
	}

	if len(validPoints) == 0 {
		return nil
	}

	estimatedSize := 0
	for i := 1; i < len(validPoints); i++ {
		dt := validPoints[i].Duration - validPoints[i-1].Duration
		if dt > 0 {
			estimatedSize += dt
		} else {
			estimatedSize++
		}
	}

	points := make([][]any, 0, estimatedSize+1)
	points = append(points, gpsPointToSample(&validPoints[0], workout, speedMS(&validPoints[0])))

	for i := 1; i < len(validPoints); i++ {
		prev := &validPoints[i-1]
		next := &validPoints[i]
		dt := next.Duration - prev.Duration
		if dt <= 0 {
			points = append(points, gpsPointToSample(next, workout, speedMS(next)))
			continue
		}

		segmentSpeed := speedMS(next)

		for step := 1; step <= dt; step++ {
			if step == dt {
				points = append(points, gpsPointToSample(next, workout, segmentSpeed))
				continue
			}

			interp := interpolateGpsPoint(prev, next, step)
			points = append(points, gpsPointToSample(&interp, workout, segmentSpeed))
		}
	}

	return points
}

// interpolateGpsPoint returns a synthetic point step seconds after prev on the way to next
func interpolateGpsPoint(prev, next *models.WorkoutDetails, step int) models.WorkoutDetails {
	dt := next.Duration - prev.Duration
	fraction := float64(step) / float64(dt)

	elapsed := time.Duration(step) * time.Second

	return models.WorkoutDetails{
		Duration:           prev.Duration + step,
		Time:               prev.Time.Add(elapsed),
		Latitude:           prev.Latitude + fraction*(next.Latitude-prev.Latitude),
		Longitude:          prev.Longitude + fraction*(next.Longitude-prev.Longitude),
		HeartRate:          prev.HeartRate,
		HorizontalAccuracy: prev.HorizontalAccuracy,
	}
}

// speedMpsForPoint returns the traveling speed in m/s
func speedMS(point *models.WorkoutDetails) any {
	if point.Speed > 0 {
		return 1000.0 / float64(point.Speed)
	}

	return nil
}

func gpsPointToSample(point *models.WorkoutDetails, workout *models.Workout, speed any) []any {
	hr := 0.0
	if point.HeartRate.Valid {
		hr = float64(point.HeartRate.Int64)
	}

	var accuracy any
	if point.HorizontalAccuracy.Valid {
		accuracy = point.HorizontalAccuracy.Float64
	}

	return []any{
		workoutPointTimeMs(point, workout),
		point.Latitude,
		point.Longitude,
		speed,
		hr,
		accuracy,
	}
}

func workoutPointTimeMs(point *models.WorkoutDetails, workout *models.Workout) int {
	if workout == nil {
		return point.Duration * 1000
	}

	return int(point.Time.Sub(workout.Start).Milliseconds())
}

func (a *Api) IsAlreadyUploaded(workout *models.Workout) (uploaded bool, href string) {
	return workout.PumpfoilorgSync.String != "", workout.PumpfoilorgSync.String
}
