package parser

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"github.com/tkrajina/gpxgo/gpx"
)

var (
	ErrCity = errors.NewError("Failed to determine nearest city", 500)
)

// Duration in seconds after which a workout is recognized as "paused"
const WorkoutPausedDiff = 60

// Resting heart rate used for "Default" calories
const RestingHeartRate = 70

// workoutParser is an internal wrapper around the parser logic
type workoutParser struct {
	user  *models.User
	input []models.GpxPoint

	// The index of the input data that is currently processed
	current value

	// The last that was processed
	last value

	// Average values
	avg avgValue

	// Maximum values
	max maxValue

	// Caluclated data to return
	rtc []models.WorkoutDetails

	// The sum of all PAI scores within the last week
	paiSumScore int
}

// value contains data in comparison to the last point
type value struct {
	// Index within the input data
	index int

	heartRate int

	// elevation in meters
	elevation int

	// 3D distance in meters from the last point
	distance int64

	// Speed in sec/km
	speed int64

	// Workout duration without pauses in seconds
	duration int64

	// Time of this point
	time time.Time

	// Sum of pai activity score
	pai float64

	lat  float64
	long float64
}

// avgValue is used to store the median values for data that needs
// to be upsampled
type avgValue struct {

	// How many data points were already processed
	count int64

	speed     float64
	heartRate float64
}

// maxValue contains maximum values used as stats and other
// purposes
type maxValue struct {
	heartRate int

	distance    int
	distanceLat float64
	distanceLon float64
}

// Internal wrapper around geonames with additional distance to point
type geonamesDistance struct {
	models.Geonames

	Distance float64
}

func (v value) ToDetails() models.WorkoutDetails {
	rtc := models.WorkoutDetails{
		Speed:     int(v.speed),
		Elevation: v.elevation,
		Latitude:  v.lat,
		Longitude: v.long,
		Duration:  int(v.duration),
		Distance:  int(v.distance),
		Time:      v.time,
	}

	// Add heart rate
	if v.heartRate != 0 {
		rtc.HeartRate = sql.NullInt64{Valid: true, Int64: int64(v.heartRate)}
	}

	return rtc
}

func newValueFromGpxPoint(point models.GpxPoint, index int) value {
	return value{
		index:     index,
		heartRate: point.HeartRate,
		elevation: point.Elevation,
		lat:       float64(point.Lat),
		long:      float64(point.Lon),
		time:      point.Timestamp,
		distance:  0,
		speed:     0,
		duration:  0,
	}
}

// Workout parses the provided (GPX) workout file and returns the
// calculated and downsampled file data to store in the database.
// Any workout metadata won't be set inside this function.
//
// You have to provide a user for calculating data like calories
func Workout(workout *models.GpxFile, user *models.User, db *database.DatabaseUtils, paiScore int) (*models.Workout, errors.Error) {
	parser := &workoutParser{
		user:  user,
		input: workout.Points,
	}

	rtc := &models.Workout{
		TypeId: workout.Type,
		UserId: user.Id,
		Start:  workout.Points[0].Timestamp,
		End:    workout.Points[len(workout.Points)-1].Timestamp,
	}

	// Parse all points
	var avg avgValue
	var max maxValue
	rtc.WorkoutDetails, avg, max = parser.Parse()

	// Fill header data we already have
	lastDetails := rtc.WorkoutDetails[len(rtc.WorkoutDetails)-1]
	rtc.Duration = lastDetails.Duration
	if avg.heartRate > 20 {
		rtc.HeartRateAv = database.NewNullInt(int(math.Round(avg.heartRate)))
		rtc.HeartRateMax = database.NewNullInt(max.heartRate)

		// Calculate calories
		rtc.Calories = CalculateBurnedCalories(rtc.Duration, int(rtc.HeartRateAv.Int64), user)
		rtc.CaloriesDefault = CalculateBurnedCalories(rtc.Duration, RestingHeartRate, user)
	}
	rtc.Distance = lastDetails.Distance
	rtc.Pai = int(math.Round(finishPaiCalculation(parser.last.pai)))

	// We cannot use the calculated average speed.
	// Because more time was spent on driving slower, we would need to do
	// a weighted average based on time and distance.
	// So we just calculate the averade time based on duration
	speedAv := float64(rtc.Duration) / (float64(rtc.Distance) / 1000.0)
	rtc.SpeedAv = int(math.Round(speedAv))

	// Calculate elevation
	rtc.ElevationUp, rtc.ElevationDown = getElevation(&rtc.WorkoutDetails)

	// Get the nearest city based on the starting point
	start := rtc.WorkoutDetails[0]
	city, err := getNearestCity(start.Longitude, start.Latitude, 20000, db)
	if err != nil {
		logger.Warning("Could not get nearest city: %s", err)
	} else {
		rtc.Country = city.Country
		rtc.City = city.Name
		rtc.CityId = city.Geonameid
		rtc.CityLocation = city.Location
	}

	return rtc, nil
}

// Parse parses all input values and returns the workoutDetails to
// store inside the database.
//
// It does also calculate some average and maximum values needed
// for the workout header
func (p *workoutParser) Parse() ([]models.WorkoutDetails, avgValue, maxValue) {

	for i, point := range p.input {

		// Initiate the data
		if i == 0 {
			p.current = newValueFromGpxPoint(point, i)
			p.last.time = p.current.time.Add(-6 * time.Second)
			p.rtc = append(p.rtc, p.current.ToDetails())

			// Calculate avg and max
			p.calcAvg()
			p.calcMax()

			// Set last
			p.last = p.current

			continue
		}

		// We don't calculate any data if the workout was stopped
		if p.wasPaused(i) {
			// Copy data from last point
			newCurrent := newValueFromGpxPoint(point, i)
			newCurrent.pai = p.last.pai
			p.last = newCurrent

			// Sum up data values we need to count up like duration and distance
			newCurrent.duration += int64(p.rtc[len(p.rtc)-1].Duration) + 1
			newCurrent.distance += int64(p.rtc[len(p.rtc)-1].Distance)

			p.rtc = append(p.rtc, newCurrent.ToDetails())
			continue
		}

		// Don't calculate data in downsample process.
		// If this is the last point before a pause, we also process it
		if !p.shouldProcess(i) && !p.wasPaused(i+1) {
			continue
		}

		// Get moving average of all values
		newCurrent := p.movingAverage(i)
		p.current = newCurrent

		// Sum up data values we need to count up like duration and distance
		newCurrent.duration += int64(p.rtc[len(p.rtc)-1].Duration)
		newCurrent.distance += int64(p.rtc[len(p.rtc)-1].Distance)

		// Calculate average
		p.calcAvg()

		// Calculate max
		p.calcMax()

		// Add value to rtc
		p.rtc = append(p.rtc, newCurrent.ToDetails())
		p.last = p.current
	}

	return p.rtc, p.avg, p.max
}

// wasPaused returns wheather a pause was made between the last and current
// point. This is true if no point was tracked during this duration
func (p *workoutParser) wasPaused(index int) bool {
	if index >= len(p.input) {
		return false
	}

	return p.input[index].Timestamp.Unix()-p.last.time.Unix() >= WorkoutPausedDiff
}

// shouldProcess returns wheather this point has to be proceeded or if we
// can skip this point to downsample the data
func (p *workoutParser) shouldProcess(index int) bool {

	// Last point is always added
	if index == len(p.input)-1 {
		return true
	}

	// Process a datapoint every six seconds
	return p.input[index].Timestamp.Unix()-p.last.time.Unix() >= 6
}

// movingAverage calculates the moving average of generic data
// like elevation and heart rate and fills the value with these
// points
func (p *workoutParser) movingAverage(index int) value {
	current := p.input[index]

	// Get last points within two seconds
	minDate := current.Timestamp.Add(-2 * time.Second).Add(-10 * time.Millisecond)
	maxDate := current.Timestamp.Add(2 * time.Second).Add(10 * time.Millisecond)

	prev := []models.GpxPoint{}
	for i := index - 1; i > 0; i-- {
		if p.input[i].Timestamp.After(minDate) {
			prev = append(prev, p.input[i])
		} else {
			break
		}
	}

	after := []models.GpxPoint{}
	for i := index + 1; i < len(p.input); i++ {
		if p.input[i].Timestamp.Before(maxDate) {
			after = append(after, p.input[i])
		} else {
			break
		}
	}

	// Sum all values
	all := prev
	all = append(all, current)
	all = append(all, after...)

	// And calculate average
	heartrates := []int{}
	elevation := []int{}

	// Build all data
	for _, pp := range all {
		heartrates = append(heartrates, pp.HeartRate)
		elevation = append(elevation, pp.Elevation)
	}

	// Add with average
	rtc := value{
		index:     index,
		heartRate: avgInt(heartrates...),
		elevation: avgInt(elevation...),
		lat:       float64(current.Lat),
		long:      float64(current.Lon),
		time:      current.Timestamp,
	}
	rtc.duration = rtc.time.Unix() - p.last.time.Unix()

	// @TODO We cannot use the average distance. This would return incorrect data
	// because we use the last point (not the "average" point) for the next calculation
	// of the distance.
	// If we would really want to build a moving average for GPX points, the Geometric median
	// should be used (e.g. Weiszfeld algorithm). So we epect the GPX device to return correct data
	// with a good GPS fix
	dist := gpx.Distance3D(
		p.last.lat, p.last.long, *gpx.NewNullableFloat64(float64(p.last.elevation)),
		float64(current.Lat), float64(current.Lon), *gpx.NewNullableFloat64(float64(rtc.elevation)),
		false,
	)
	rtc.distance = int64(math.Round((dist)))

	// Calculate speed in seconds/km (based on average values)
	if rtc.duration != 0 && rtc.distance > 0 {
		speed := float64(1000) / (float64(rtc.distance) / float64(rtc.duration))
		rtc.speed = int64(math.Round(speed))
	}

	// Sum of pai activity score
	rtc.pai = p.last.pai + calculateAcitivityScore(int(rtc.duration), rtc.heartRate, p.paiSumScore, p.user)

	return rtc
}

// calcAvg adds the [current] values to the average state based
// on the pastime since [last] point.
//
// It does calculate missing data with a linear graph between [last]
// and [current] if there are no 6s steps available.
//
// Because there can be many data points available, the incremental update
// method is used instead of storing the sum
func (p *workoutParser) calcAvg() {
	timePast := p.current.time.Unix() - p.last.time.Unix()

	// Simple incremental step
	if timePast <= 6 {
		for i := 1; i <= int(timePast); i++ {
			p.avg.count++

			val := float64(p.current.heartRate)
			p.avg.heartRate += (val - p.avg.heartRate) / float64(p.avg.count)
			val = float64(p.current.speed)
			p.avg.speed += (val - p.avg.speed) / float64(p.avg.count)
		}

		return
	}

	// Draw a vector between last and current point and calculate value at specific time
	stepsBasis := float64(p.current.heartRate - p.last.heartRate)
	step := stepsBasis / float64(timePast)
	for i := 1; i <= int(timePast); i++ {
		p.avg.count++

		val := step*float64(i) + float64(p.last.heartRate)
		p.avg.heartRate += (val - p.avg.heartRate) / float64(p.avg.count)
		val = step*float64(i) + float64(p.last.speed)
		p.avg.speed += (val - p.avg.speed) / float64(p.avg.count)
	}
}

// calcMax updates the maximum values if [current] exceeds the previously
// maximum values
func (p *workoutParser) calcMax() {

	// Heart rate
	if p.current.heartRate > p.max.heartRate {
		p.max.heartRate = p.current.heartRate
	}

	// Farest away point
	dist := gpx.Distance2D(
		p.rtc[0].Latitude, p.rtc[0].Longitude,
		p.current.lat, p.current.long,
		false,
	)
	if dist > float64(p.max.distance) {
		p.max.distance = int(math.Round(dist))
	}

}

// CalculateBurnedCalories calculates the burned calories for the given workout
// duration and average heart rate.
//
// The calculation is based on the v02max value of the provided user
func CalculateBurnedCalories(duration, avgHeartRate int, usr *models.User) int {
	var calories float64
	if usr.Gender == models.GENDER_MALE {
		calories = (float64(duration) / 60) * (0.6309*float64(avgHeartRate) + 0.1988*float64(usr.Weight) + 0.2017*float64(time.Now().Year()-usr.BirthYear) - 55.0906) / 4.184
	} else {
		calories = (float64(duration) / 60) * (0.4472*float64(avgHeartRate) + 0.1263*float64(usr.Weight) + 0.074*float64(time.Now().Year()-usr.BirthYear) - 20.4022) / 4.184
	}

	return int(math.Round(calories))
}

// getElevation calculates the traveled elevation up- and downhill
// during the provided workout
func getElevation(data *[]models.WorkoutDetails) (up, down int) {
	lastIndex := 0

	for i, p := range *data {
		if i == 0 {
			continue
		}

		// Only calculate the difference every 30 seconds
		if p.Duration-(*data)[lastIndex].Duration < 30 {
			continue
		}

		diff := p.Elevation - (*data)[lastIndex].Elevation
		if diff > 0 {
			up += diff
		} else {
			down += diff
		}
		lastIndex = i
	}

	return
}

// getNearestCity returns the nearest bigger city to the given
// location based on GeoDB data read from db in te provided radius
func getNearestCity(lon, lat float64, radius int, db *database.DatabaseUtils) (rtc models.Geonames, err error) {
	if db == nil {
		logger.Debug("No database provided in getNearestCity")
		return
	}

	// Get bounds to improve performance
	lonMin, lonMax, latMin, latMax := GetBoundingBox(lon, lat, radius)

	// Build select
	geonames := []geonamesDistance{}
	sel := db.Struct.QuerySlice(&geonames)

	// Use a quadrat to improve performance (is more efficent than calculating radius for every row)
	// Use st_makeEnvelope (point(?, ?),point(?, ?)),lonMin, latMin, lonMax, latMax if mariadb supports it!
	sel.Where().Custom(`
		ST_Contains(
			ST_GeomFromText(CONCAT( 'POLYGON((', ?, ' ', ?, ', ', ?, ' ', ?, ', ', ?, ' ',  ?, ', ', ?, ' ', ?,  ', ', ?, ' ', ?, '))')),
			location
		)
	`, lonMin, latMin, lonMin, latMax, lonMax, latMax, lonMax, latMin, lonMin, latMin).Add()

	// Get exact radius to point
	sel.CustomColumn("", "Distance", fmt.Sprintf(
		`ST_Distance_Sphere(location, point(%f, %f)) AS "distance"`,
		lon, lat,
	))
	sel.Where().Custom("( ST_Distance_Sphere(location, point(?, ?)) ) <= ?", lon, lat, radius).Add()
	sel.OrderBy("", "distance", "ASC")

	if err := sel.Run(); err != nil {
		return models.Geonames{}, err
	}

	// Get nearest and biggest city
	sort.SliceStable(geonames, func(aa, bb int) bool {
		a := geonames[aa]
		b := geonames[bb]
		return getCityWeight(a.Distance, a.Population) < getCityWeight(b.Distance, b.Population)
	})

	// Use the first city
	if len(geonames) > 0 {
		rtc = geonames[0].Geonames
	}

	// No city found → increase search radius
	if rtc.Geonameid == 0 && radius < 60000 {
		return getNearestCity(lon, lat, 60000, db)
	} else if rtc.Geonameid == 0 {
		return models.Geonames{}, fmt.Errorf("no city in radius of 60km found")
	}

	return rtc, nil
}

// getCityWeight returns the weight of a city with the provided details
func getCityWeight(distance float64, population int) int {
	popMultiplier := float64(population) / 5000.0
	// Set boundings
	if popMultiplier > 3 {
		popMultiplier = 3.0
	} else if popMultiplier < 1 {
		popMultiplier = 1.0
	}

	return int(math.Round(distance / popMultiplier))
}

// GetBoundingBox returns the bounds of a square which contains all points
// that are inside a circly starting from center with the provided radius
// in meters.
//
// It's used to improve query performance for mysql
func GetBoundingBox(lon, lat float64, radius int) (boundLonMin, boundLonMax, boundLatMin, boundLatMax float64) {

	// Increase radius of 50 meters for rounding :)
	radius += 50

	_, boundLonMin = AddMetersToPosition(lat, lon, float64(radius*-1), true)
	_, boundLonMax = AddMetersToPosition(lat, lon, float64(radius), true)

	boundLatMin, _ = AddMetersToPosition(lat, lon, float64(radius*-1), false)
	boundLatMax, _ = AddMetersToPosition(lat, lon, float64(radius), false)

	return
}

func avg(vals ...int) float64 {
	var total int = 0
	for _, value := range vals {
		total += value
	}

	return float64(total) / float64(len(vals))
}

func avgInt(vals ...int) int {
	return int(math.Round(avg(vals...)))
}

// calculateAcitivityScore calculates a physical acitivity indicator based on the workout
// duration and heartrate.
//
// Because the score is NOT liniear to the heart rate, you should calculate the score
// every minute and sum the returning values up.
//
// This function tries to replicate the propritary PAI score
func calculateAcitivityScore(duration int, heartRate int, paiWeek int, user *models.User) float64 {
	if heartRate < 20 {
		return 0
	}

	// @TODO use resting heart rate from user
	min := 60
	age := time.Now().Year() - user.BirthYear

	hrr := 220.0 - (float64(age) * 0.9) - float64(min)
	if user.Gender == models.GENDER_FEMALE {
		hrr = 206.0 - (0.86 * float64(age))
	}

	// Minimum workout intensivity to start "earning" pai
	minIntensivity := 25.0

	intensivity := ((float64(heartRate) - float64(min)) / hrr) * 100
	paiMin := 0.0004 * ((float64(intensivity) - minIntensivity) * (float64(intensivity) - minIntensivity + 8))
	paiSum := paiMin * (float64(duration) / 60.0)

	// Accumulate minItensivity
	if intensivity < minIntensivity || paiSum < 0 {
		paiSum = 0
	}

	// It's harder to earn pai over time
	if paiWeek > 100 {
		paiSum *= 0.8
	} else if paiWeek > 50 {
		paiSum *= 0.9
	} else if paiWeek > 20 {
		paiSum *= 0.95
	}

	return paiSum
}

// finishPaiCalculation finishes the calculation of the whole PAIs
// earned by a workout
func finishPaiCalculation(paiSum float64) float64 {
	if paiSum > 70 {
		paiSum = paiSum * 0.7
	} else if paiSum > 50 {
		paiSum = paiSum * 0.8
	} else if paiSum > 20 {
		paiSum = paiSum * 0.9
	}

	// Ignore very small pais
	if paiSum <= 1 {
		return 0
	}

	return paiSum
}
