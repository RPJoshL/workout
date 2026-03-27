package parser

import (
	"fmt"
	"math"
	"testing"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/dbutils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/tests"
	"git.rpjosh.de/RPJosh/workout/pkg/assert"
	"github.com/RPJoshL/go-ddl-parser"
	"github.com/RPJoshL/go-logger"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/guregu/null/v5"
)

// TestParserSecond tests the parsing of workout data that
// was secondly tracked
func TestParserSecond(t *testing.T) {
	input := []models.GpxPoint{
		{
			Timestamp: timeWithOffset(0),
			Elevation: 400,
			HeartRate: 100,
			Lat:       AddMetersToBaseLat(0, 0),
			Lon:       AddMetersToBaseLon(0, 0),
			Steps:     5,
		},
		{
			Timestamp: timeWithOffset(1),
			Elevation: 999,
			HeartRate: 100,
			Lat:       AddMetersToBaseLat(0, 0),
			Lon:       AddMetersToBaseLon(0, 0),
			Steps:     5,
		},
		{
			Timestamp: timeWithOffset(2),
			Elevation: 999,
			HeartRate: 100,
			Lat:       AddMetersToBaseLat(0, 0),
			Lon:       AddMetersToBaseLon(0, 0),
			Steps:     10,
		},
		{
			Timestamp: timeWithOffset(3),
			Elevation: 999,
			HeartRate: 100,
			Lat:       AddMetersToBaseLat(0, 0),
			Lon:       AddMetersToBaseLon(0, 0),
			Steps:     10,
		},
		// >>>>>>>>>> AVG - Test average of Heartrate, Elevation and 3d Point
		{
			Timestamp: timeWithOffset(4),
			Elevation: 412,
			HeartRate: 120,
			Lat:       AddMetersToBaseLat(0, 0),
			Lon:       AddMetersToBaseLon(0, 0),
			Steps:     12,
		},
		{
			Timestamp: timeWithOffset(5),
			Elevation: 418,
			HeartRate: 122,
			Lat:       AddMetersToBaseLat(0, 0),
			Lon:       AddMetersToBaseLon(0, 0),
			Steps:     14,
		},
		{
			Timestamp: timeWithOffset(6),
			Elevation: 420,
			HeartRate: 130,
			Lat:       AddMetersToBaseLat(0, 0),
			Lon:       AddMetersToBaseLon(0, 0),
			Steps:     16,
		},
		{
			Timestamp: timeWithOffset(7),
			Elevation: 415,
			HeartRate: 122,
			Lat:       AddMetersToBaseLat(0, 0),
			Lon:       AddMetersToBaseLon(0, 0),
			Steps:     20,
		},
		{
			Timestamp: timeWithOffset(8),
			Elevation: 418,
			HeartRate: 128,
			Lat:       AddMetersToBaseLat(0, 0),
			Lon:       AddMetersToBaseLon(0, 0),
			Steps:     22,
		},
		// <<<<<<<<<<<< AVG
		{
			Timestamp: timeWithOffset(9),
			Elevation: 417,
			Lat:       AddMetersToBaseLat(0, 0),
			Lon:       AddMetersToBaseLon(0, 0),
			Steps:     25,
		},
		// >>>>>>>>>> AVG - Test linear points
		{
			Timestamp: timeWithOffset(10),
			Elevation: 417,
			Lat:       AddMetersToBaseLat(10, 0),
			Lon:       AddMetersToBaseLon(10, 0),
			Steps:     30,
		},
		{
			Timestamp: timeWithOffset(11),
			Elevation: 417,
			Lat:       AddMetersToBaseLat(20, 0),
			Lon:       AddMetersToBaseLon(20, 0),
			Steps:     33,
		},
		{
			Timestamp: timeWithOffset(12),
			Elevation: 417,
			Lat:       AddMetersToBaseLat(30, 0),
			Lon:       AddMetersToBaseLon(30, 0),
			Steps:     33,
		},
		{
			Timestamp: timeWithOffset(13),
			Elevation: 417,
			Lat:       AddMetersToBaseLat(40, 0),
			Lon:       AddMetersToBaseLon(40, 0),
			Steps:     33,
		},
		// <<<<<<<<<<<< AVG
	}

	expected := []models.WorkoutDetails{
		{
			Duration:  0,
			Elevation: 400,
			Latitude:  float64(input[0].Lat),
			Longitude: float64(input[0].Lon),
			HeartRate: null.IntFrom(100),
			Time:      timeWithOffset(0),
			StepCount: null.IntFrom(5),
		},
		{
			Duration:  6,
			Elevation: 417,
			Distance:  17,
			Speed:     353,
			Latitude:  float64(input[6].Lat),
			Longitude: float64(input[6].Lon),
			HeartRate: null.IntFrom(124),
			Time:      timeWithOffset(6),
			StepCount: null.IntFrom(16),
		},
		{
			Duration:  12,
			Elevation: 417,
			Distance:  47, // 17 + 30
			Speed:     200,
			Latitude:  float64(input[12].Lat),
			Longitude: float64(input[12].Lon),
			Time:      timeWithOffset(12),
			StepCount: null.IntFrom(33),
		},
		{
			Duration:  13,
			Elevation: 417,
			Distance:  57, // 17 + 30 + 10
			Speed:     100,
			Latitude:  float64(input[13].Lat),
			Longitude: float64(input[13].Lon),
			Time:      timeWithOffset(13),
			StepCount: null.IntFrom(33),
		},
	}

	got, err := Workout(&models.GpxFile{Points: input}, &models.User{
		Gender: models.GENDER_MALE,
		Weight: 70,
		Height: 178,
		Vo2Max: 54,
	}, nil, 0)
	if err != nil {
		t.Errorf("Failed to parse workout: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(expected, got.WorkoutDetails); diff != "" {
		t.Errorf("Mismatch of parser (-want +got):\n%s", diff)
	}
}

// TestPause tests the parsing of workout with a pause in it
func TestPause(t *testing.T) {
	input := []models.GpxPoint{
		{
			Timestamp: timeWithOffset(0),
			HeartRate: 100,
			Lat:       AddMetersToBaseLat(0, 0),
			Lon:       AddMetersToBaseLon(0, 0),
		},
		// >>>>>>>>> AVG
		{
			Timestamp: timeWithOffset(1),
			HeartRate: 100,
			Lat:       AddMetersToBaseLat(10, 0),
			Lon:       AddMetersToBaseLon(10, 0),
		},
		{
			Timestamp: timeWithOffset(2),
			HeartRate: 120,
			Lat:       AddMetersToBaseLat(20, 0),
			Lon:       AddMetersToBaseLon(20, 0),
		},
		// <<<<<<<<<<< AVG
		{
			Timestamp: timeWithOffset(200),
			HeartRate: 80,
			Lat:       AddMetersToBaseLat(200, 0),
			Lon:       AddMetersToBaseLon(200, 0),
		},
		{
			Timestamp: timeWithOffset(201),
			HeartRate: 100,
			Lat:       AddMetersToBaseLat(210, 0),
			Lon:       AddMetersToBaseLon(210, 0),
		},
	}

	expected := []models.WorkoutDetails{
		{
			Duration:  0,
			Latitude:  float64(input[0].Lat),
			Longitude: float64(input[0].Lon),
			HeartRate: null.IntFrom(100),
			Time:      timeWithOffset(0),
		},
		{
			Duration:  2,
			Distance:  20,
			Speed:     100,
			Latitude:  float64(input[2].Lat),
			Longitude: float64(input[2].Lon),
			HeartRate: null.IntFrom(110),
			Time:      timeWithOffset(2),
		},
		// Pause
		{
			Duration:  3,
			Distance:  20,
			Speed:     0,
			Latitude:  float64(input[3].Lat),
			Longitude: float64(input[3].Lon),
			HeartRate: null.IntFrom(80),
			Time:      timeWithOffset(200),
		},
		{
			Duration:  4,
			Distance:  30, // 20 + 10
			Speed:     100,
			Latitude:  float64(input[4].Lat),
			Longitude: float64(input[4].Lon),
			HeartRate: null.IntFrom(90),
			Time:      timeWithOffset(201),
		},
	}

	got, err := Workout(&models.GpxFile{Points: input}, &models.User{
		Gender: models.GENDER_MALE,
		Weight: 70,
		Height: 178,
		Vo2Max: 54,
	}, nil, 0)
	if err != nil {
		t.Errorf("Failed to parse workout: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(expected, got.WorkoutDetails); diff != "" {
		t.Errorf("Mismatch of parser (-want +got):\n%s", diff)
	}
}

// TestAverage tests the calculation of average workout data
// with regula data points every seconds
func TestAverageSimple(t *testing.T) {
	input := []models.GpxPoint{
		{Timestamp: timeWithOffset(0), HeartRate: 105},
		{Timestamp: timeWithOffset(1), HeartRate: 0},
		{Timestamp: timeWithOffset(2), HeartRate: 0},
		{Timestamp: timeWithOffset(3), HeartRate: 0},
		{Timestamp: timeWithOffset(4), HeartRate: 130},
		{Timestamp: timeWithOffset(5), HeartRate: 170},
		{Timestamp: timeWithOffset(6), HeartRate: 145},
		{Timestamp: timeWithOffset(7), HeartRate: 100},
		{Timestamp: timeWithOffset(8), HeartRate: 80},
		{Timestamp: timeWithOffset(9), HeartRate: 190},
		{Timestamp: timeWithOffset(10), HeartRate: 144},
	}

	// Expected values (we test only the heart rate).
	// Calculating the speed should be the same!
	expectedAvgHeartRate := int(math.Round((105*6 + 125*6 + 138*4) / float64(16)))

	got, err := Workout(&models.GpxFile{Points: input}, &models.User{}, nil, 0)
	if err != nil {
		t.Errorf("Failed to parse workout: %s", err)
	}

	if int(got.HeartRateAv.Int64) != expectedAvgHeartRate {
		t.Errorf("Missmatch of avg heart rate. Expected %d. Got %d", expectedAvgHeartRate, got.HeartRateAv.Int64)
	}
}

// TestAverageNotPeriodic tests the calculation of average workout data
// with unregular data points with steps > 6 seconds
func TestAverageNotPeriodic(t *testing.T) {
	input := []models.GpxPoint{
		{Timestamp: timeWithOffset(0), HeartRate: 105},
		{Timestamp: timeWithOffset(20), HeartRate: 200},
		{Timestamp: timeWithOffset(26), HeartRate: 170},
		{Timestamp: timeWithOffset(27), HeartRate: 160},
		{Timestamp: timeWithOffset(40), HeartRate: 70},
		{Timestamp: timeWithOffset(41), HeartRate: 80},
	}

	// Expected values (we test only the heart rate).
	// These values were determined in GeoGebra
	expectedAvgHeartRate := int(math.Round((105*6 + 133.5*6 + 162*6 + 190.5*6 + 200*2 + 165*6 + 125.25*6 + 84*6 + 75*3) / float64(47)))

	got, err := Workout(&models.GpxFile{Points: input}, &models.User{}, nil, 0)
	if err != nil {
		t.Errorf("Failed to parse workout: %s", err)
	}

	if int(got.HeartRateAv.Int64) != expectedAvgHeartRate {
		t.Errorf("Missmatch of avg heart rate. Expected %d. Got %d", expectedAvgHeartRate, got.HeartRateAv.Int64)
	}
}

func TestBoundingBox(t *testing.T) {
	radius := 5000
	centerLat := 48.3939
	centerLon := 10.5355

	aLat, aLon := AddMetersToPosition(centerLat, centerLon, float64(radius), false)
	logger.Debug("5km -> LAT %f / LON %f", aLat, aLon)

	lonMin, lonMax, latMin, latMax := GetBoundingBox(centerLon, centerLat, radius)
	if lonMin > centerLon || lonMax < centerLon || latMin > centerLat || latMax < centerLat {
		t.Errorf("Received incorrect bounds: | LAT < %f AND LAT > %f | LON < %f AND LON > %f", latMax, latMin, lonMax, lonMin)
	} else {
		logger.Debug("LAT < %f AND LAT > %f | LON < %f AND LON > %f", latMax, latMin, lonMax, lonMin)
	}
}

func TestNearestCity(t *testing.T) {
	db := dbutils.NewByDb(tests.GetDbConnection(t))

	data := []models.Geonames{
		{
			Geonameid:      2923439,
			Name:           "Gablingen",
			Alternatenames: null.StringFrom("Gablingen,jia bu lin gen"),
			Location:       ddl.Location{Longitude: 10.81667, Latitude: 48.45},
			Country:        "DE",
			Population:     4707,
		},
		{
			Geonameid:      2920891,
			Name:           "Gersthofen",
			Alternatenames: null.StringFrom("Gerstgofen,Gersthofen,Gerstkhofen"),
			Location:       ddl.Location{Longitude: 10.87273, Latitude: 48.42432},
			Country:        "DE",
			Population:     20254,
		},
		{
			Geonameid:      2764957,
			Name:           "Sölden",
			Alternatenames: null.StringFrom("Soelden,Soeldeni vald,Solden,Sölden,Söldeni vald"),
			Location:       ddl.Location{Longitude: 11, Latitude: 46.96667},
			Country:        "AT",
			Population:     2205,
		},
		{
			Geonameid:      8740442,
			Name:           "Längenfeld",
			Alternatenames: null.StringFrom("Lengenfel'd,lun gen fu er de,rengenferuto,Ленгенфельд"),
			Location:       ddl.Location{Longitude: 10.96951, Latitude: 47.07398},
			Country:        "AT",
			Population:     4611,
		},
	}

	_, err := db.Struct.InsertSlice(&data).Run()
	assert.NoError(t, err)

	// City that is nearer to "Gablingen", but "Gesthofen" is bigger
	centerLat := 48.44048
	centerLon := 10.83020
	city, dbErr := GetNearestCity(centerLon, centerLat, 20000, db)
	assert.NoError(t, dbErr)
	assert.Equal(t, "Gersthofen", city.Name)

	// "Sölden" should be used because "Längenfeld" is not that much bigger
	centerLat = 46.96924
	centerLon = 10.86281
	city, dbErr = GetNearestCity(centerLon, centerLat, 20000, db)
	assert.NoError(t, dbErr)
	assert.Equal(t, "Sölden", city.Name)
}

// getTimeWithOffset returns a time with the added "offsetSeconds" to
// a constant date and time
func timeWithOffset(offsetSeconds int) time.Time {
	baseTime, err := time.Parse("2006-01-02T15:04:05Z", "2024-05-02T11:20:00Z")
	if err != nil {
		panic(fmt.Sprintf("Failed to parse base time: %s", err))
	}

	return baseTime.Add(time.Duration(offsetSeconds) * time.Second)
}

// TestVerticalExtremes tests the removal of horizontal exremes (elevation) of input points
func TestVerticalExtremes(t *testing.T) {
	p := workoutParser{
		input: []models.GpxPoint{
			{Timestamp: timeWithOffset(0), Elevation: 400},
			{Timestamp: timeWithOffset(6), Elevation: 401},
			{Timestamp: timeWithOffset(12), Elevation: 400},
			{Timestamp: timeWithOffset(18), Elevation: 402},
			{Timestamp: timeWithOffset(24), Elevation: 0}, // Use last elevation
			{Timestamp: timeWithOffset(32), Elevation: 0}, // Use last elevation
			{Timestamp: timeWithOffset(40), Elevation: 399},
			{Timestamp: timeWithOffset(46), Elevation: 400},
			{Timestamp: timeWithOffset(52), Elevation: 305}, // Extreme => replace
			{Timestamp: timeWithOffset(60), Elevation: 280}, // Extreme => replace
			{Timestamp: timeWithOffset(66), Elevation: 387},
			{Timestamp: timeWithOffset(72), Elevation: 390},
			{Timestamp: timeWithOffset(78), Elevation: 392},
			{Timestamp: timeWithOffset(84), Elevation: 395},
		},
	}

	// Correct points
	p.preparePoints()

	// Expected points
	expected := []models.GpxPoint{
		{Timestamp: timeWithOffset(0), Elevation: 400},
		{Timestamp: timeWithOffset(6), Elevation: 401},
		{Timestamp: timeWithOffset(12), Elevation: 400},
		{Timestamp: timeWithOffset(18), Elevation: 402},
		{Timestamp: timeWithOffset(24), Elevation: 402}, // Use last elevation
		{Timestamp: timeWithOffset(32), Elevation: 402}, // Use last elevation
		{Timestamp: timeWithOffset(40), Elevation: 399},
		{Timestamp: timeWithOffset(46), Elevation: 400},
		{Timestamp: timeWithOffset(52), Elevation: 400}, // Extreme => replace
		{Timestamp: timeWithOffset(60), Elevation: 400}, // Extreme => replace
		{Timestamp: timeWithOffset(66), Elevation: 387},
		{Timestamp: timeWithOffset(72), Elevation: 390},
		{Timestamp: timeWithOffset(78), Elevation: 392},
		{Timestamp: timeWithOffset(84), Elevation: 395},
	}

	// Compare structs
	if diff := cmp.Diff(expected, p.input); diff != "" {
		t.Errorf("Mismatch of removing the vertical extremes (-want +got):\n%s", diff)
	}
}

// TestOverwriteDeviceData tests the overwriting of calculated data with
// data that was tracked by the device
func TestOverwriteDeviceData(t *testing.T) {
	input := []models.GpxPoint{
		{Timestamp: timeWithOffset(0), Speed: 200, Distance: 0},
		{Timestamp: timeWithOffset(1), Speed: 200, Distance: 200},
		{Timestamp: timeWithOffset(2), Speed: 200, Distance: 400},
		{Timestamp: timeWithOffset(3), Speed: 300, Distance: 700},
		{Timestamp: timeWithOffset(4), Speed: 300, Distance: 1000}, // avg (1)
		{Timestamp: timeWithOffset(5), Speed: 300, Distance: 1300}, // avg (1)
		{Timestamp: timeWithOffset(6), Speed: 500, Distance: 1800}, // avg (1)
		{Timestamp: timeWithOffset(7), Speed: 500, Distance: 2300}, // avg (1)
		{Timestamp: timeWithOffset(8), Speed: 500, Distance: 2800}, // avg (1)
		{Timestamp: timeWithOffset(9), Speed: 400, Distance: 3200},
		{Timestamp: timeWithOffset(10), Speed: 400, Distance: 3600}, // avg (2)
		{Timestamp: timeWithOffset(11), Speed: 600, Distance: 4000}, // avg (2)
		{Timestamp: timeWithOffset(12), Speed: 800, Distance: 4600}, // avg (2)
	}

	// We use data that does not match the calculated data from above to see that hard
	// overwrite does work
	deviceData := models.DeviceData{
		UseDeviceData: true,
		SpeedAvg:      1000,
		DistanceTotal: 5000,
	}

	got, err := Workout(&models.GpxFile{Points: input, DeviceData: deviceData}, &models.User{}, nil, 0)
	if err != nil {
		t.Errorf("Failed to parse workout: %s", err)
	}

	expected := &models.Workout{
		WorkoutDetails: []models.WorkoutDetails{
			{
				Duration: 0,
				Distance: 0,
				Speed:    200,
				Time:     timeWithOffset(0),
			},
			{
				Duration: 6,
				Distance: 1800,
				Speed:    420,
				Time:     timeWithOffset(6),
			},
			{
				Duration: 12,
				Distance: 4600,
				Speed:    600,
				Time:     timeWithOffset(12),
			},
		},
		SpeedAv:  1000,
		Distance: 5000,
		Start:    timeWithOffset(0),
		End:      timeWithOffset(12),
		Duration: 12,
	}

	assert.EqualStruct(t, "Workout", expected, got)
}

// TestDistanceCalculationWithMissingGPSFix tests the distance calculation
// when an initial GPS fix is missing for the workout
func TestDistanceCalculationWithMissingGPSFix(t *testing.T) {
	input := []models.GpxPoint{
		{Timestamp: timeWithOffset(0)},
		{Timestamp: timeWithOffset(60)},
		{Timestamp: timeWithOffset(120)},
		{Timestamp: timeWithOffset(180)},
		{Timestamp: timeWithOffset(320)},
		{Timestamp: timeWithOffset(350), Lat: AddMetersToBaseLat(0, 0), Lon: AddMetersToBaseLon(0, 0)},
		{Timestamp: timeWithOffset(380), Lat: AddMetersToBaseLat(100, 0), Lon: AddMetersToBaseLon(100, 0)},
	}

	got, err := Workout(&models.GpxFile{Points: input, DeviceData: models.DeviceData{}}, &models.User{}, nil, 0)
	if err != nil {
		t.Errorf("Failed to parse workout: %s", err)
	}

	expected := &models.Workout{
		WorkoutDetails: []models.WorkoutDetails{
			{
				Duration: 0,
				Distance: 0,
				Time:     timeWithOffset(0),
			},
			{
				Duration: 60,
				Distance: 0,
				Time:     timeWithOffset(60),
			},
			{
				Duration: 120,
				Distance: 0,
				Time:     timeWithOffset(120),
			},
			{
				Duration: 180,
				Distance: 0,
				Time:     timeWithOffset(180),
			},
			{
				Duration: 181,
				Distance: 0,
				Time:     timeWithOffset(320),
			},
			{
				Duration: 211,
				Distance: 0,
				Time:     timeWithOffset(350),
			},
			{
				Duration: 241,
				Distance: 100,
				Speed:    300,
				Time:     timeWithOffset(380),
			},
		},
		Distance: 100,
		Start:    timeWithOffset(0),
		End:      timeWithOffset(380),
		Duration: 241,
		// The speed is actually not correct, because the first valid GPS point is at second 350.
		// But that should be fine in such cases
		SpeedAv: 2410,
	}

	assert.EqualStruct(
		t, "Workout", expected, got,
		cmpopts.IgnoreFields(models.WorkoutDetails{}, "Latitude", "Longitude"),
	)
}
